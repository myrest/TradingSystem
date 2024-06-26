package services

import (
	"context"
	"log"
	"os"
	"sync"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

var (
	app             *firebase.App
	firestoreMu     sync.Mutex
	firestoreClient *firestore.Client
)

func init() {
	ctx := context.Background()
	var err error
	credsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credsPath == "" {
		firestoreClient, err = firestore.NewClient(ctx, "resttradingsystem")
		if err != nil {
			//log.Fatalf("error initializing Firestore client: %v\n", err)
			log.Println("error initializing Firestore client: \n", err.Error())
		}
	} else {
		sa := option.WithCredentialsFile(credsPath)

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
}

func getFirestoreClient() *firestore.Client {
	firestoreMu.Lock()
	defer firestoreMu.Unlock()
	return firestoreClient
}
