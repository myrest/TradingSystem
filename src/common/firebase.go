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

// FirebaseManager 封裝 Firebase 相關操作
type FirebaseManager struct {
	app             *firebase.App
	firestoreClient *firestore.Client
	config          *FirebaseConfig
	mu              sync.RWMutex
	initialized     bool
}

// FirebaseConfig 包含 Firebase 配置
type FirebaseConfig struct {
	ProjectID         string
	FireBaseKeyPath   string
	OAuthKeyPath      string
	SecretManagerPath string
	UseSecretManager  bool
}

type firebaseConfigFile struct {
	APIKey            string `json:"apiKey"`
	AuthDomain        string `json:"authDomain"`
	ProjectID         string `json:"projectId"`
	StorageBucket     string `json:"storageBucket"`
	MessagingSenderID string `json:"messagingSenderId"`
	AppID             string `json:"appId"`
}

var (
	firebaseManager *FirebaseManager
	once            sync.Once
	secManagerCache string
	secManagerMu    sync.RWMutex
)

// GetFirebaseManager 返回 Firebase 管理器的單例實例
func GetFirebaseManager() *FirebaseManager {
	once.Do(func() {
		firebaseManager = &FirebaseManager{}
	})
	return firebaseManager
}

// Initialize 初始化 Firebase
func (fm *FirebaseManager) Initialize(ctx context.Context, env EnviromentType) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if fm.initialized {
		return nil
	}

	config, err := fm.buildConfig(env)
	if err != nil {
		return fmt.Errorf("failed to build config: %w", err)
	}

	fm.config = config

	if err := fm.initializeFirebaseApp(ctx); err != nil {
		return fmt.Errorf("failed to initialize Firebase app: %w", err)
	}

	if err := fm.initializeFirestore(ctx); err != nil {
		return fmt.Errorf("failed to initialize Firestore: %w", err)
	}

	fm.initialized = true
	return nil
}

// buildConfig 建立 Firebase 配置
func (fm *FirebaseManager) buildConfig(env EnviromentType) (*FirebaseConfig, error) {
	root := os.Getenv("KEYROOT")
	if root == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
		root = filepath.Dir(wd)
	}

	envStr := env.String()
	oauthPath := filepath.Join(root, fmt.Sprintf("firebaseConfig_%s.json", envStr))
	firebasePath := filepath.Join(root, fmt.Sprintf("serviceAccountKey_%s.json", envStr))

	// 讀取 OAuth 配置文件以獲取項目信息
	configFile, err := fm.loadConfigFile(oauthPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load OAuth config file: %w", err)
	}

	if err := fm.validateConfigFile(configFile); err != nil {
		return nil, fmt.Errorf("invalid config file: %w", err)
	}

	secretPath := fmt.Sprintf("projects/%s/secrets/GOOGLE_APPLICATION_CREDENTIALS/versions/latest",
		configFile.MessagingSenderID)

	return &FirebaseConfig{
		ProjectID:         configFile.ProjectID,
		FireBaseKeyPath:   firebasePath,
		OAuthKeyPath:      oauthPath,
		SecretManagerPath: secretPath,
		UseSecretManager:  !IsFileExists(firebasePath),
	}, nil
}

// loadConfigFile 從檔案載入 Firebase 配置
func (fm *FirebaseManager) loadConfigFile(filename string) (*firebaseConfigFile, error) {
	if !IsFileExists(filename) {
		return nil, fmt.Errorf("config file does not exist: %s", filename)
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	var config firebaseConfigFile
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON from %s: %w", filename, err)
	}

	return &config, nil
}

// validateConfigFile 驗證配置文件的完整性
func (fm *FirebaseManager) validateConfigFile(config *firebaseConfigFile) error {
	if config.ProjectID == "" {
		return fmt.Errorf("projectId is required")
	}
	if config.MessagingSenderID == "" {
		return fmt.Errorf("messagingSenderId is required")
	}
	return nil
}

// initializeFirebaseApp 初始化 Firebase App
func (fm *FirebaseManager) initializeFirebaseApp(ctx context.Context) error {
	var sa option.ClientOption
	var err error

	if fm.config.UseSecretManager {
		creds, err := fm.GetSecretFromManager(ctx)
		if err != nil {
			return fmt.Errorf("failed to get credentials from Secret Manager: %w", err)
		}
		sa = option.WithCredentialsJSON([]byte(creds))
	} else {
		sa = option.WithCredentialsFile(fm.config.FireBaseKeyPath)
	}

	fm.app, err = firebase.NewApp(ctx, nil, sa)
	if err != nil {
		return fmt.Errorf("failed to initialize Firebase app: %w", err)
	}

	return nil
}

// initializeFirestore 初始化 Firestore 客戶端
func (fm *FirebaseManager) initializeFirestore(ctx context.Context) error {
	var err error
	fm.firestoreClient, err = fm.app.Firestore(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize Firestore client: %w", err)
	}
	return nil
}

// getSecretFromManager 從 Google Secret Manager 獲取密鑰
func (fm *FirebaseManager) GetSecretFromManager(ctx context.Context) (string, error) {
	secManagerMu.RLock()
	if secManagerCache != "" {
		defer secManagerMu.RUnlock()
		return secManagerCache, nil
	}
	secManagerMu.RUnlock()

	secManagerMu.Lock()
	defer secManagerMu.Unlock()

	// 雙重檢查
	if secManagerCache != "" {
		return secManagerCache, nil
	}

	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create Secret Manager client: %w", err)
	}
	defer client.Close()

	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fm.config.SecretManagerPath,
	}

	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to access secret version: %w", err)
	}

	secManagerCache = string(result.Payload.Data)
	return secManagerCache, nil
}

// GetFirestoreClient 返回 Firestore 客戶端
func (fm *FirebaseManager) GetFirestoreClient() (*firestore.Client, error) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	if !fm.initialized {
		return nil, fmt.Errorf("firebase manager not initialized")
	}

	if fm.firestoreClient == nil {
		return nil, fmt.Errorf("firestore client not available")
	}

	return fm.firestoreClient, nil
}

// GetAuth 返回 Firebase Auth 客戶端
func (fm *FirebaseManager) GetAuth(ctx context.Context) (*auth.Client, error) {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	if !fm.initialized {
		return nil, fmt.Errorf("firebase manager not initialized")
	}

	if fm.app == nil {
		return nil, fmt.Errorf("firebase app not available")
	}

	return fm.app.Auth(ctx)
}

// Close 關閉 Firebase 連接
func (fm *FirebaseManager) Close() error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if fm.firestoreClient != nil {
		if err := fm.firestoreClient.Close(); err != nil {
			return fmt.Errorf("failed to close Firestore client: %w", err)
		}
		fm.firestoreClient = nil
	}

	fm.app = nil
	fm.initialized = false
	return nil
}

// GetConfig 返回 Firebase 配置
func (fm *FirebaseManager) GetConfig() *FirebaseConfig {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return fm.config
}

// IsInitialized 檢查是否已初始化
func (fm *FirebaseManager) IsInitialized() bool {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return fm.initialized
}

// 向後兼容的函數
func GetFirestoreClient() *firestore.Client {
	client, err := GetFirebaseManager().GetFirestoreClient()
	if err != nil {
		// 為了向後兼容，如果出錯就返回 nil
		// 在實際應用中，應該要處理這個錯誤
		return nil
	}
	return client
}

func FirebaseAuth(ctx context.Context) (*auth.Client, error) {
	return GetFirebaseManager().GetAuth(ctx)
}

// InitializeFirebase 初始化 Firebase（新的公共函數）
func InitializeFirebase(ctx context.Context, env EnviromentType) error {
	return GetFirebaseManager().Initialize(ctx, env)
}

// 向後兼容的結構和函數
type FirebaseVariable struct {
	FireBaseKeyFullPath string `firestore:"-"` //連Firebase的key，不能放DB裏
	OAuthKeyFullPath    string `firestore:"-"` //連OAuth的key
	ProjectID           string `firestore:"-"` //Firebase專案ID
}

// GetFirebaseSetting 向後兼容函數，返回 FirebaseVariable
func GetFirebaseSetting() FirebaseVariable {
	manager := GetFirebaseManager()
	config := manager.GetConfig()

	if config == nil {
		// 如果還沒初始化，返回空的結構
		return FirebaseVariable{}
	}

	return FirebaseVariable{
		FireBaseKeyFullPath: config.FireBaseKeyPath,
		OAuthKeyFullPath:    config.OAuthKeyPath,
		ProjectID:           config.ProjectID,
	}
}

// GetSecret 向後兼容函數，從 Secret Manager 獲取密鑰
func GetSecret(ctx context.Context) (string, error) {
	manager := GetFirebaseManager()
	if !manager.IsInitialized() {
		return "", fmt.Errorf("firebase manager not initialized")
	}
	return manager.GetSecretFromManager(ctx)
}
