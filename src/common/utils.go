package common

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type SystemSettings struct {
	FireBaseKeyFullPath string
	OAuthKeyFullPath    string
	Env                 EnviromentType
}

type EnviromentType string

const (
	Prod EnviromentType = "prod"
	Dev  EnviromentType = "dev"
)

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

func GetEnvironmentSetting() SystemSettings {
	var rtn SystemSettings
	wd, _ := os.Getwd()
	if err := godotenv.Load(filepath.Join(wd, ".env")); err != nil {
		log.Printf("No .env file. Use system environment variable.")
	}
	root := os.Getenv("KEYROOT")
	env := os.Getenv("ENVIRONMENT")

	if env == "" || strings.ToLower(env) == "prod" {
		rtn.Env = Prod
	} else {
		rtn.Env = Dev
	}
	rtn.OAuthKeyFullPath = filepath.Join(root, fmt.Sprintf("firebaseConfig_%s.json", rtn.Env))
	rtn.FireBaseKeyFullPath = filepath.Join(root, fmt.Sprintf("serviceAccountKey_%s.json", rtn.Env))

	return rtn
}
