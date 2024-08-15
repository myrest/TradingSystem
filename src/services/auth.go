package services

import (
	"TradingSystem/src/models"
	"context"
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

func VerifyIDTokenAndGetDetails(idToken string) (models.GoogleTokenDetail, error) {
	var rtn models.GoogleTokenDetail
	ctx := context.Background()
	authClient, err := app.Auth(ctx)
	if err != nil {
		return rtn, err
	}

	token, err := authClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		return rtn, err
	}

	uid := token.UID
	email, _ := token.Claims["email"].(string)
	name, _ := token.Claims["name"].(string)
	photo, _ := token.Claims["picture"].(string)

	rtn = models.GoogleTokenDetail{
		UID:   uid,
		Email: email,
		Name:  name,
		Photo: photo,
	}

	return rtn, nil
}
