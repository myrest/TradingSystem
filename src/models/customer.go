package models

import "TradingSystem/src/common"

type Customer struct {
	ID                 string              `json:"id"`
	Name               string              `json:"name"`
	Email              string              `json:"email"`
	APIKey             string              `json:"apikey"`
	SecretKey          string              `json:"secretkey"`
	IsAdmin            bool                `json:"isadmin"`
	IsAutoSubscribe    bool                `json:"autosubscribe"`
	AutoSubscribReal   bool                `json:"subscribtype"`
	AutoSubscribAmount int                 `json:"amount"`
	TgChatID           int64               `json:"tgchatid"`
	TgIdentifyKey      string              `json:"tgidentifykey"`
	AlertMessageType   AlertMessageModel   `json:"alertmessagetype"`
	ExchangeSystemName ExchangeSystem      `json:"exchangesystem"`
	DataCenter         common.ServerLocale `json:"dataCenter"`
}

type CustomerRelationUI struct {
	Parent_CustomerID string
	Parent_Email      string
	Parent_Name       string
	Customer          Customer
}

type AlertMessageModel string
type ExchangeSystem string

const (
	CustomerAlertAll     AlertMessageModel = "All"               //全都發
	CustomerAlertClose   AlertMessageModel = "Close"             //平倉發
	CustomerAlertLoss    AlertMessageModel = "Loss"              //虧損發
	CustomerAlertDefault AlertMessageModel = "Default"           //預設，下單失敗、開第六倉及日結通知
	ExchangeBingx        ExchangeSystem    = "Bingx"             //Bingx
	ExchangeBinance_N    ExchangeSystem    = "Binance_Normal"    //Binance 一般帳戶
	ExchangeBinance_P    ExchangeSystem    = "Binance_Portfolio" //Binance 統一帳戶
)

// 定義一個警告消息的結構體
type alertMessage struct {
	Name     AlertMessageModel `json:"name"`
	Priority int               `json:"priority"`
}

var alertMessages = []alertMessage{
	{Name: CustomerAlertAll, Priority: 3},
	{Name: CustomerAlertClose, Priority: 2},
	{Name: CustomerAlertLoss, Priority: 1},
	{Name: CustomerAlertDefault, Priority: 0},
}

// 為 AlertMessageModel 添加 GetPriority 方法
func (a *AlertMessageModel) GetPriority() int {
	for _, alert := range alertMessages {
		if alert.Name == *a {
			return alert.Priority
		}
	}
	return -1 // 未知的類型
}
