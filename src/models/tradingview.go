package models

import (
	"TradingSystem/src/bingx"
	"TradingSystem/src/common"
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
	Price  string `json:"price"`
	Cert   string `json:"cert"`
	Symbol string `json:"symbol"`
	Time   string `json:"time"`
}

type CustomerCurrencySymboWithCustomer struct {
	CustomerCurrencySymbol
	Customer
}

type TVData struct {
	Action       string
	Contracts    float64
	PositionSize float64
	Price        float64
	Symbol       string
}

// 程式內用的TV訊號值
type TvSiginalData struct {
	TVData
	PlaceOrderType
}

type PlaceOrderType struct {
	PositionSideType bingx.PositionSideType //Long, Short
	Side             bingx.SideType         //Buy, Sell
}

type Log_TvSiginalData struct {
	PlaceOrderType
	Profit       float64
	CustomerID   string
	Result       string
	Time         string
	Amount       float64
	Price        float64
	Simulation   bool
	WebHookRefID string
	Symbol       string
}

// 依訊號來決定倉位及方向
func (t *TvSiginalData) Convert(d TvWebhookData) {
	t.TVData.Action = strings.ToLower(d.Data.Action)
	t.TVData.Contracts, _ = strconv.ParseFloat(d.Data.Contracts, 64)
	t.TVData.PositionSize, _ = strconv.ParseFloat(d.Data.PositionSize, 64)
	t.TVData.Price, _ = strconv.ParseFloat(d.Price, 64)
	t.TVData.Symbol = common.FormatSymbol(d.Symbol)
	if t.TVData.PositionSize == 0 {
		//平倉
		if t.TVData.Action == "sell" {
			t.PlaceOrderType.Side = bingx.SellSideType
			t.PlaceOrderType.PositionSideType = bingx.LongPositionSideType //平多倉
		} else {
			t.PlaceOrderType.Side = bingx.BuySideType
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
				t.PlaceOrderType.PositionSideType = bingx.LongPositionSideType
			}
		} else {
			//應持空倉
			if t.TVData.Action == "sell" {
				//開空倉 or 加空倉
				t.PlaceOrderType.Side = bingx.SellSideType
				t.PlaceOrderType.PositionSideType = bingx.ShortPositionSideType
			} else {
				//減倉
				t.PlaceOrderType.Side = bingx.BuySideType
				t.PlaceOrderType.PositionSideType = bingx.ShortPositionSideType
			}
		}
	}
}
