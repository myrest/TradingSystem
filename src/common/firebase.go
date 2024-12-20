package common

import (
	"context"
	"log"
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

func GetFirestoreClient() *firestore.Client {
	firestoreMu.Lock()
	defer firestoreMu.Unlock()
	return firestoreClient
}
