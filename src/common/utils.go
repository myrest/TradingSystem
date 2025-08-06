package common

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/joho/godotenv"
)

type SystemSettings struct {
	Env             EnviromentType `firestore:"Env"`            //環境
	DemoCustomerID  string         `firestore:"DemoCustomerID"` //測試用的CustomerID
	TempCacheFolder string         `firestore:"-"`              //緩存資料夾 Todo:目前沒有使用
	TgToken         string         `firestore:"TgToken"`        //Telegram Token
	StartTimestemp  string         `firestore:"-"`              //開始時間，用來避免Static文件被瀏覽器快取
	SectestWord     string         `firestore:"-"`              //後門路由
}

var systemSettingsLock sync.Mutex // 定義一個互斥鎖

// region 環境
const (
	EmptyEnvironment EnviromentType = iota // 預設為空值
	Dev                                    // http://localhost:8080/
	Prod                                   // https://hikari.lolo.finance/
	GoogleJP                               // https://trading.innoroot.com/
)

type EnviromentType int

// 將字串轉換為 EnviromentType
func StringToEnviromentType(s string) (EnviromentType, bool) {
	switch strings.ToLower(s) {
	case "prod":
		return Prod, false
	case "googlejp":
		return GoogleJP, false
	case "dev":
		return Dev, false
	default:
		return EmptyEnvironment, true // 返回一個錯誤
	}
}

// 使用 stringer 生成的 String() 方法
func (e EnviromentType) String() string {
	return [...]string{"EmptyEnvironment", "dev", "prod", "googlejp"}[e]
}

// endregion 環境

var systemSettings SystemSettings

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func loadEnvironmentFromFile() bool {
	//判斷執行目錄是否有.env檔案
	if IsFileExists(".env") {
		//如果有 .env檔就載入
		wd, _ := os.Getwd()
		if err := godotenv.Load(filepath.Join(wd, ".env")); err == nil {
			log.Printf("使用.env檔，作為環境變數")
		}
		return true
	}
	return false
}

func getTempCacheFolder() string {
	//tempCacheFolder固定為"tmpCache"
	tmpFolderName := "tmpCache"
	wd, _ := os.Getwd()
	return filepath.Join(wd, tmpFolderName)
}

func GenerateRandomString(length int) string {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// Convert "USDT.P" to "-USDT"
func FormatSymbol(symbol string, DashFlag ...bool) string {
	addDash := true
	if len(DashFlag) > 0 {
		addDash = DashFlag[0]
	}
	if addDash {
		return regexp.MustCompile(`USDT\.P`).ReplaceAllString(symbol, "-USDT")
	} else {
		return regexp.MustCompile(`USDT\.P`).ReplaceAllString(symbol, "USDT")
	}
}

func IsFileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// 因為很常被呼叫，所以統一寫在這
func GetEnvironmentSetting() SystemSettings {
	if systemSettings.Env != EmptyEnvironment {
		return systemSettings
	}

	systemSettingsLock.Lock()         // 獲取鎖
	defer systemSettingsLock.Unlock() // 確保在函數結束時釋放鎖

	if systemSettings.Env != EmptyEnvironment {
		return systemSettings
	}

	//優先權為
	//1. .env檔，因為會將檔案載入，變成環境變數
	//2. 機器上的環境變數
	//最後會一律套用DB setting，目前只有支援DemoCustomerID及TgToken
	loadEnvironmentFromFile() //讀取.env檔的環境變數

	strEnv := os.Getenv("ENVIRONMENT")

	env, err := StringToEnviromentType(strEnv)
	if err {
		panic("環境變數不正確")
	}

	systemSettings = SystemSettings{
		Env:             env,
		TempCacheFolder: getTempCacheFolder(),
		StartTimestemp:  strconv.FormatInt(time.Now().Unix(), 10),
		SectestWord:     GenerateRandomString(8),
		TgToken:         "", //由DB取得
		DemoCustomerID:  "", //由DB取得
	}

	//初始化DB，因為會取systemSettings，所以要放在後面
	if err := InitializeFirebase(context.Background(), env); err != nil {
		panic(fmt.Errorf("failed to initialize Firebase: %w", err))
	}
	dbSystemSetting, _ := GetDBSystemSettings(context.Background())
	ApplySystemSettings(dbSystemSetting)

	return systemSettings
}

func ApplySystemSettings(settings SystemSettings) {
	systemSettings.DemoCustomerID = settings.DemoCustomerID
	systemSettings.TgToken = settings.TgToken
}

func Decimal(value interface{}, rounds ...int) float64 {
	round := 8
	if rounds != nil {
		round = rounds[0]
	}
	format := fmt.Sprintf("%%.%df", round)
	switch value.(type) {
	case float64:
		rtn, _ := strconv.ParseFloat(fmt.Sprintf(format, value), 64)
		return rtn
	case float32:
		rtn, _ := strconv.ParseFloat(fmt.Sprintf(format, value), 64)
		return rtn
	default:
		rtn, _ := strconv.ParseFloat(fmt.Sprintf("%s", value), 64)
		return rtn
	}
}

func GetReportStartEndDate(s sessions.Session) (time.Time, time.Time) {
	sdt := s.Get("report_sdt")
	edt := s.Get("report_edt")
	if sdt != nil && edt != nil {
		sddate := FormatDate(ParseTime(sdt.(string)))
		eddate := fmt.Sprintf("%s 23:59:59", FormatDate(ParseTime(edt.(string))))
		return ParseTime(sddate), ParseTime(eddate)
	}
	return TimeMax(), TimeMax()
}

func SetReportStartEndDate(s sessions.Session, sdt, edt time.Time) {
	s.Set("report_sdt", FormatDate(sdt))
	s.Set("report_edt", fmt.Sprintf("%s 23:59:59", FormatDate(edt)))
	_ = s.Save() //不處理失敗
}

func IsEmail(email string) bool {
	// Regular expression pattern to validate email addresses
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

	// Using regexp package to match the email against the pattern
	matched, err := regexp.MatchString(pattern, email)
	if err != nil {
		return false
	}

	return matched
}
