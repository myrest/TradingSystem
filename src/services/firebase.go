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
	log.Printf("啟動 (%s) 中..", "Firebase")
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
		if app == nil {
			log.Fatalln("firebase.NewApp got empty.")
		} else {
			log.Println("firebase.NewApp is good.")
		}

		// Initialize Firestore client
		firestoreClient, err = app.Firestore(ctx)
		if err != nil {
			log.Fatalf("error initializing Firestore client: %v\n", err)
		}
		if firestoreClient == nil {
			log.Fatalln("app.Firestore got empty.")
		} else {
			log.Println("app.Firestore is good.")
		}
	}
}

func getFirestoreClient() *firestore.Client {
	firestoreMu.Lock()
	defer firestoreMu.Unlock()
	return firestoreClient
}
