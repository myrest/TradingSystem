package common

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"cloud.google.com/go/firestore"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

var (
	app             *firebase.App
	firestoreMu     sync.Mutex
	firestoreClient *firestore.Client
)
var secmanagerCert string
var firebaseVariable FirebaseVariable
var firebaseconfigfile firebaseConfigFile

type FirebaseVariable struct {
	FireBaseKeyFullPath string `firestore:"-"` //連Firebase的key，不能放DB裏
	OAuthKeyFullPath    string `firestore:"-"` //連OAuth的key
	ProjectID           string `firestore:"-"` //Firebase專案ID
}

type firebaseConfigFile struct {
	APIKey            string `json:"apiKey"`
	AuthDomain        string `json:"authDomain"`
	ProjectID         string `json:"projectId"`
	StorageBucket     string `json:"storageBucket"`
	MessagingSenderID string `json:"messagingSenderId"`
	AppID             string `json:"appId"`
}

func initFirebaseSetting() {
	//先手動取得OAuth及Firebase Config，因為後面有需要共用
	root := os.Getenv("KEYROOT") //這個會是由外部變數來決定ServiceAccount放哪裏
	//沒有設定Key的目錄，就以當前執行目錄為設定檔目錄
	if root == "" {
		wd, _ := os.Getwd()
		root = filepath.Dir(wd)
	}
	OAuthKeyFullPath := filepath.Join(root, fmt.Sprintf("firebaseConfig_%s.json", systemSettings.Env.String()))
	FireBaseKeyFullPath := filepath.Join(root, fmt.Sprintf("serviceAccountKey_%s.json", systemSettings.Env.String()))

	//完成firebaseconfigfile -> OAuth
	firebaseconfigfile = initFirebaseConfigFile(OAuthKeyFullPath)

	//完成firebaseconfigfile -> Firebase
	firebaseVariable = FirebaseVariable{
		FireBaseKeyFullPath: FireBaseKeyFullPath,
		OAuthKeyFullPath:    OAuthKeyFullPath,
		ProjectID:           firebaseconfigfile.ProjectID,
	}
	//初始化Firebase Database
	initFirebaseDatabase()
}

func initFirebaseDatabase() {
	ctx := context.Background()
	var err error

	// Initialize Firebase，如果設定的檔案不存在，就從Google Secret Manager 取得
	var sa option.ClientOption
	if IsFileExists(firebaseVariable.FireBaseKeyFullPath) {
		sa = option.WithCredentialsFile(firebaseVariable.FireBaseKeyFullPath)
	} else {
		creds, err := GetSecret(ctx)
		if err != nil {
			panic(fmt.Errorf("failed to access secret version: %v", err))
		}
		sa = option.WithCredentialsJSON([]byte(creds))
	}

	app, err = firebase.NewApp(ctx, nil, sa)
	if err != nil {
		panic(fmt.Errorf("error initializing firebase.NewApp(): %v", err))
	}

	// Initialize Firestore client
	firestoreClient, err = app.Firestore(ctx)
	if err != nil {
		panic(fmt.Errorf("error initializing Firestore client app.Firestore(): %v", err))
	}
}

// 從檔案取得Firebase Config，用來做Oauth使用
func initFirebaseConfigFile(filename string) (rtn firebaseConfigFile) {
	// Open the JSON file
	file, err := os.Open(filename)
	if err != nil {
		panic(fmt.Errorf("failed to open file: %w", err))
	}
	defer file.Close()

	// Read the file contents
	data, err := io.ReadAll(file)
	if err != nil {
		panic(fmt.Errorf("failed to read file: %w", err))
	}

	// Parse the JSON data
	if err := json.Unmarshal(data, &rtn); err != nil {
		panic(fmt.Errorf("failed to unmarshal JSON: %w", err))
	}
	return rtn
}

func FirebaseAuth(ctx context.Context) (*auth.Client, error) {
	return app.Auth(ctx)
}

func GetSecret(ctx context.Context) (string, error) {
	if secmanagerCert != "" {
		return secmanagerCert, nil
	}

	//"projects/635522974118/secrets/GOOGLE_APPLICATION_CREDENTIALS/versions/latest"
	securityFullPath := fmt.Sprintf("projects/%s/secrets/GOOGLE_APPLICATION_CREDENTIALS/versions/latest", firebaseconfigfile.MessagingSenderID)

	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()

	accessRequest := &secretmanagerpb.AccessSecretVersionRequest{
		Name: securityFullPath,
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

func GetFirebaseSetting() FirebaseVariable {
	return firebaseVariable
}
