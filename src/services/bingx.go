package services

import (
	"ManageAPI/src/bingx"
	"context"
	"strconv"
)

func GetAccountBalance(APIkey, SecretKey string) (float64, error) {
	client := bingx.NewClient(APIkey, SecretKey)
	res, err := client.NewGetBalanceService().Do(context.Background())
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat((*res).AavailableMargin, 64)
}
