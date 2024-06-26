package models

import (
	"TradingSystem/src/bingx"
	"strconv"
	"strings"
)

// 從TV收到的資料
type TvWebhookData struct {
	Data struct {
		Action       string `json:"action"`
		Contracts    string `json:"contracts"`
		PositionSize string `json:"position_size"`
	} `json:"data"`
	Price    string `json:"price"`
	UserInfo string `json:"user_info"`
	Symbol   string `json:"symbol"`
	Time     string `json:"time"`
}

type CustomerCurrencySymboWithCustomer struct {
	CustomerCurrencySymbo
	Customer
}

// 程式內用的TV訊號值
type TvSiginalData struct {
	TVData struct {
		Action       string
		Contracts    float64
		PositionSize float64
		Price        float64
		Symbol       string
	}
	PlaceOrderType struct {
		PositionSideType bingx.PositionSideType //Long, Short
		Side             bingx.SideType         //Buy, Sell
	}
}

// 依訊號來決定倉位及方向
func (t *TvSiginalData) Convert(d TvWebhookData) {
	t.TVData.Action = strings.ToLower(d.Data.Action)
	t.TVData.Contracts, _ = strconv.ParseFloat(d.Data.Contracts, 64)
	t.TVData.PositionSize, _ = strconv.ParseFloat(d.Data.PositionSize, 64)
	t.TVData.Price, _ = strconv.ParseFloat(d.Price, 64)
	t.TVData.Symbol = formatSymbol(d.Symbol)
	if t.TVData.PositionSize == 0 {
		//平倉
		t.PlaceOrderType.Side = bingx.SellSideType
		if t.TVData.Action == "sell" {
			t.PlaceOrderType.PositionSideType = bingx.LongPositionSideType //平多倉
		} else {
			t.PlaceOrderType.PositionSideType = bingx.ShortPositionSideType //平空倉
		}
	} else {
		//可能為開倉、加倉，減倉
		//先判斷是買or賣
		if t.TVData.PositionSize > 0 {
			//應持多倉
			if t.TVData.Action == "buy" {
				//開多倉 or 加多倉
				t.PlaceOrderType.Side = bingx.BuySideType
				t.PlaceOrderType.PositionSideType = bingx.LongPositionSideType
			} else {
				//減倉
				t.PlaceOrderType.Side = bingx.SellSideType
				t.PlaceOrderType.PositionSideType = bingx.ShortPositionSideType
			}
		} else {
			//應持空倉
			if t.TVData.Action == "sell" {
				//開空倉 or 加空倉
				t.PlaceOrderType.Side = bingx.BuySideType
				t.PlaceOrderType.PositionSideType = bingx.ShortPositionSideType
			} else {
				//減倉
				t.PlaceOrderType.Side = bingx.SellSideType
				t.PlaceOrderType.PositionSideType = bingx.LongPositionSideType
			}
		}
	}
}

func formatSymbol(symbol string) string {
	// Split the symbol into base and currency parts
	parts := strings.Split(symbol, "USDT.P")
	if len(parts) != 2 {
		return symbol
	}

	// Format the symbol with "-" before "USDT"
	formattedSymbol := parts[0] + "-" + "USDT"
	return formattedSymbol
}
