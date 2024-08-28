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

type CustomerWeeklyReport struct {
	YearWeek   string //週數 YYYY-MM
	CustomerID string //客戶ID
	DemoSymbolList
}

func (e CustomerWeeklyReport) Merge(obj CustomerWeeklyReport) {
	e.Amount += obj.Amount
	e.CloseCount += obj.CloseCount
	e.LossCount += obj.LossCount
	e.OpenCount += obj.OpenCount
	e.Profit += obj.Profit
	e.WinCount += obj.WinCount
	e.Winrate += obj.Winrate
}

type CustomerWeeklyReportSummary struct {
	YearWeek string  //週數 YYYY-MM
	Profit   float64 //獲利
}
