package models

type SubAccount struct {
	SubAccountDB
	DocumentRefID string `json:"refid"` //Firebase Ref ID
}

type SubAccountDB struct {
	CustomerID  string //本尊ID
	AccountName string `json:"accountname"` //帳號名稱
}

type SubAccountUI struct {
	AccountName   string `json:"accountname"` //帳號名稱
	DocumentRefID string `json:"refid"`       //Firebase Ref ID
}
