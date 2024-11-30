package services

import (
	"TradingSystem/src/common"
	"TradingSystem/src/models"
	"context"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"google.golang.org/api/iterator"
)

func TGSendMessage(chatid int64, message string) error {
	settings := common.GetEnvironmentSetting()
	bot, err := tgbotapi.NewBotAPI(settings.TgToken)
	if err != nil {
		log.Panic(err)
	}

	msg := tgbotapi.NewMessage(chatid, message)
	_, err = bot.Send(msg)
	if err != nil {
		return err
	}
	return nil
}

func SetTGIdentifyKey(ctx context.Context, CustomerID string) (string, error) {
	customer, err := GetCustomer(ctx, CustomerID)
	if err != nil {
		return "", err
	}
	customer.TgIdentifyKey = common.GenerateRandomString(10)
	err = UpdateCustomer(ctx, customer)
	if err != nil {
		return "", err
	}
	return customer.TgIdentifyKey, nil
}

func GetCustomerByTGChatIDss(ctx context.Context, ChatID int64) (*models.Customer, error) {
	client := common.GetFirestoreClient()

	iter := client.Collection("customers").Where("TgChatID", "==", ChatID).Limit(1).Documents(ctx)
	defer iter.Stop()
	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, nil // Customer not found
	}
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var customer models.Customer
	doc.DataTo(&customer)
	customer.ID = doc.Ref.ID
	return &customer, nil
}
