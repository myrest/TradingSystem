package models

type DemoSymbolList struct {
	Symbol     string  //幣種
	Profit     float64 //獲利
	CloseCount int32   //平倉次數
	OpenCount  int32   //開倉次數
	WinCount   int32   //盈利次數
	LossCount  int32   //虧損次數
	Amount     float64 //交易額
	Winrate    string  //勝率
}
