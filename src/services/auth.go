package services

import (
	"context"
	"log"
)

func VerifyIDToken(idToken string) (string, error) {
	ctx := context.Background()
	authClient, err := app.Auth(ctx)
	if err != nil {
		return "", err
	}

	token, err := authClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		return "", err
	}

	return token.UID, nil
}

func VerifyIDTokenAndGetDetails(idToken string) (string, string, string, error) {
	ctx := context.Background()
	if app == nil {
		log.Println("VerifyIDTokenAndGetDetails中的app 為NULL.")
	} else {
		log.Println("VerifyIDTokenAndGetDetails執行中")
	}
	authClient, err := app.Auth(ctx)
	if err != nil {
		log.Printf("VerifyIDTokenAndGetDetails執行有問題\n %s", err.Error())
		return "", "", "", err
	}

	token, err := authClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		log.Printf("出錯了:%s", err)
		return "", "", "", err
	}

	uid := token.UID
	email, _ := token.Claims["email"].(string)
	name, _ := token.Claims["name"].(string)

	return uid, email, name, nil
}
