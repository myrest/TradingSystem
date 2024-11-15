package common

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/joho/godotenv"
)

type FirebaseSettings struct {
	FireBaseKeyFullPath string `firestore:"-"` //連Firebase的key，不能放DB裏
	OAuthKeyFullPath    string `firestore:"-"` //連OAuth的key
	ProjectID           string `firestore:"-"` //Firebase專案ID
}

type SystemSettings struct {
	Env             EnviromentType `firestore:"Env"`            //環境
	DemoCustomerID  string         `firestore:"DemoCustomerID"` //測試用的CustomerID
	TempCacheFolder string         `firestore:"-"`              //緩存資料夾
	TgToken         string         `firestore:"TgToken"`        //Telegram Token
	StartTimestemp  string         `firestore:"-"`              //開始時間，用來避免Static文件被瀏覽器快取
	SectestWord     string         `firestore:"-"`              //後門路由
}

const (
	Prod EnviromentType = "prod"
	Dev  EnviromentType = "dev"
)

type firebaseConfig struct {
	APIKey            string `json:"apiKey"`
	AuthDomain        string `json:"authDomain"`
	ProjectID         string `json:"projectId"`
	StorageBucket     string `json:"storageBucket"`
	MessagingSenderID string `json:"messagingSenderId"`
	AppID             string `json:"appId"`
}

type EnviromentType string

var systemSettings SystemSettings
var firebaseSettings FirebaseSettings

func DecodeGzip(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	decodedMsg, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return decodedMsg, nil
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateRandomString(length int) string {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// Convert "USDT.P" to "-USDT"
func FormatSymbol(symbol string) string {
	return regexp.MustCompile(`USDT\.P`).ReplaceAllString(symbol, "-USDT")
}

func IsFileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func GetFirebaseSetting() FirebaseSettings {
	if firebaseSettings.FireBaseKeyFullPath != "" {
		return firebaseSettings
	}

	//載入.env當作環境變數，如果有成功，要顯示訊息。
	wd, _ := os.Getwd()
	if err := godotenv.Load(filepath.Join(wd, ".env")); err == nil {
		log.Printf("使用.env檔，作為環境變數")
	}

	root := os.Getenv("KEYROOT")
	env := os.Getenv("ENVIRONMENT") //用來取得Prod or Dev的檔案資料
	var rtn FirebaseSettings

	//沒有設定Key的目錄，就以當前執行目錄為設定檔目錄
	if root == "" {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Error getting current working directory: %v", err)
		}
		root = filepath.Dir(wd)
	}

	//預設為Prod
	Environment := Prod
	//嚴格限制只有Prod才是真的Production環境
	if strings.ToLower(env) != "prod" {
		Environment = Dev
	}

	rtn.OAuthKeyFullPath = filepath.Join(root, fmt.Sprintf("firebaseConfig_%s.json", Environment))
	rtn.FireBaseKeyFullPath = filepath.Join(root, fmt.Sprintf("serviceAccountKey_%s.json", Environment))
	projectid, err := getProjectID(rtn.OAuthKeyFullPath)

	if err != nil {
		log.Fatalf("Error getting project id: %v", err)
	}
	rtn.ProjectID = projectid

	firebaseSettings = rtn
	return rtn
}

func GetEnvironmentSetting() SystemSettings {
	if systemSettings.Env != "" {
		return systemSettings
	}

	var rtn SystemSettings

	//載入.env當作環境變數
	wd, _ := os.Getwd()
	if err := godotenv.Load(filepath.Join(wd, ".env")); err == nil {
		log.Printf("使用.env檔，作為環境變數")
	}

	env := os.Getenv("ENVIRONMENT")
	tmpCacheFolder := os.Getenv("TEMPCACHEFOLDER")

	if env == "" || strings.ToLower(env) == "prod" {
		rtn.Env = Prod
	} else {
		rtn.Env = Dev
	}

	rtn.TempCacheFolder = filepath.Join(wd, tmpCacheFolder)
	rtn.StartTimestemp = strconv.FormatInt(time.Now().Unix(), 10)
	rtn.SectestWord = GenerateRandomString(8)
	dbSystemSetting, _ := GetDBSystemSettings(context.Background())
	systemSettings = rtn //先把值寫入，再ApplyDB的資料
	ApplySystemSettings(dbSystemSetting)
	return rtn
}

func ApplySystemSettings(settings SystemSettings) {
	systemSettings.DemoCustomerID = settings.DemoCustomerID
	systemSettings.TgToken = settings.TgToken
}

// GetProjectID reads the firebaseConfig_dev.json file and returns the projectId value
func getProjectID(filename string) (string, error) {
	// Open the JSON file
	file, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Read the file contents
	data, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Parse the JSON data
	var config firebaseConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return config.ProjectID, nil
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
