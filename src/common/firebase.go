package common

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	"cloud.google.com/go/firestore"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

var (
	app             *firebase.App
	firestoreMu     sync.Mutex
	firestoreClient *firestore.Client
)
var secmanagerCert string
var firebaseSettings FirebaseSettings

type FirebaseSettings struct {
	FireBaseKeyFullPath string `firestore:"-"` //連Firebase的key，不能放DB裏
	OAuthKeyFullPath    string `firestore:"-"` //連OAuth的key
	ProjectID           string `firestore:"-"` //Firebase專案ID
}

type firebaseConfig struct {
	APIKey            string `json:"apiKey"`
	AuthDomain        string `json:"authDomain"`
	ProjectID         string `json:"projectId"`
	StorageBucket     string `json:"storageBucket"`
	MessagingSenderID string `json:"messagingSenderId"`
	AppID             string `json:"appId"`
}

func init() {
	ctx := context.Background()
	var err error

	settings := GetFirebaseSetting()

	var sa option.ClientOption
	if IsFileExists(settings.FireBaseKeyFullPath) {
		sa = option.WithCredentialsFile(settings.FireBaseKeyFullPath)
	} else {
		creds, err := GetSecret(ctx, "projects/635522974118/secrets/GOOGLE_APPLICATION_CREDENTIALS/versions/latest")
		if err != nil {
			log.Fatalf("failed to access secret version: %v", err)
		}
		sa = option.WithCredentialsJSON([]byte(creds))
	}

	app, err = firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	// Initialize Firestore client
	firestoreClient, err = app.Firestore(ctx)
	if err != nil {
		log.Fatalf("error initializing Firestore client: %v\n", err)
	}
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

func FirebaseAuth(ctx context.Context) (*auth.Client, error) {
	return app.Auth(ctx)
}

func GetSecret(ctx context.Context, name string) (string, error) {
	if secmanagerCert != "" {
		return secmanagerCert, nil
	}
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()

	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	}

	result, err := client.AccessSecretVersion(ctx, accessRequest)
	if err != nil {
		return "", err
	}
	secmanagerCert = string(result.Payload.Data)
	return secmanagerCert, nil
}

func GetFirestoreClient() *firestore.Client {
	firestoreMu.Lock()
	defer firestoreMu.Unlock()
	return firestoreClient
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

	root := os.Getenv("KEYROOT") //這個會是由外部變數來決定
	setting := GetEnvironmentSetting()
	var rtn FirebaseSettings

	//沒有設定Key的目錄，就以當前執行目錄為設定檔目錄
	if root == "" {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Error getting current working directory: %v", err)
		}
		root = filepath.Dir(wd)
	}

	rtn.OAuthKeyFullPath = filepath.Join(root, fmt.Sprintf("firebaseConfig_%s.json", setting.Env.String()))
	rtn.FireBaseKeyFullPath = filepath.Join(root, fmt.Sprintf("serviceAccountKey_%s.json", setting.Env.String()))
	projectid, err := getProjectID(rtn.OAuthKeyFullPath)

	if err != nil {
		log.Fatalf("Error getting project id: %v", err)
	}
	rtn.ProjectID = projectid

	firebaseSettings = rtn
	return rtn
}
