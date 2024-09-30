package models

type Customer struct {
	ID                 string            `json:"id"`
	Name               string            `json:"name"`
	Email              string            `json:"email"`
	APIKey             string            `json:"apikey"`
	SecretKey          string            `json:"secretkey"`
	IsAdmin            bool              `json:"isadmin"`
	IsAutoSubscribe    bool              `json:"autosubscribe"`
	AutoSubscribReal   bool              `json:"subscribtype"`
	AutoSubscribAmount int               `json:"amount"`
	TgChatID           int64             `json:"tgchatid"`
	TgIdentifyKey      string            `json:"tgidentifykey"`
	AlertMessageType   AlertMessageModel `json:"alertmessagetype"`
}

type CustomerRelationUI struct {
	Parent_CustomerID string
	Parent_Email      string
	Parent_Name       string
	Customer          Customer
}

type AlertMessageModel string

const (
	CustomerAlertAll     AlertMessageModel = "All"     //全都發
	CustomerAlertClose   AlertMessageModel = "Close"   //平倉發
	CustomerAlertLoss    AlertMessageModel = "Loss"    //虧損發
	CustomerAlertDefault AlertMessageModel = "Default" //預設，下單失敗、開第六倉、日結通知
)
