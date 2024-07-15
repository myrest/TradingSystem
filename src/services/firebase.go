package services

import (
	"TradingSystem/src/common"
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"cloud.google.com/go/firestore"
	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

var (
	app             *firebase.App
	firestoreMu     sync.Mutex
	firestoreClient *firestore.Client
)

func getSecret(ctx context.Context, name string) (string, error) {
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

	return string(result.Payload.Data), nil
}

func init() {
	ctx := context.Background()
	var err error

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %v", err)
	}
	firebaseKey := os.Getenv("ENVIRONMENT")
	if firebaseKey != "" && strings.ToLower(firebaseKey) == "dev" {
		firebaseKey = "dev"
	} else {
		firebaseKey = "prod"
	}

	credsPath := filepath.Join(wd, "./../serviceAccountKey_"+firebaseKey+".json")

	var sa option.ClientOption
	if common.IsFileExists(credsPath) {
		sa = option.WithCredentialsFile(credsPath)
	} else {
		creds, err := getSecret(ctx, "projects/635522974118/secrets/GOOGLE_APPLICATION_CREDENTIALS/versions/latest")
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

func getFirestoreClient() *firestore.Client {
	firestoreMu.Lock()
	defer firestoreMu.Unlock()
	return firestoreClient
}
