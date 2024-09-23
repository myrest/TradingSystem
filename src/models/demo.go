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

type CustomerProfitReport struct {
	YearUnit   string //週數 YYYY-WW 或月數YYYY-MM
	CustomerID string //客戶ID
	DemoSymbolList
}

func (e CustomerProfitReport) Merge(obj CustomerProfitReport) {
	e.Amount += obj.Amount
	e.CloseCount += obj.CloseCount
	e.LossCount += obj.LossCount
	e.OpenCount += obj.OpenCount
	e.Profit += obj.Profit
	e.WinCount += obj.WinCount
	e.Winrate += obj.Winrate
}

type CustomerReportSummary struct {
	YearUnit string  //週數 YYYY-WW 或是月數 YYYY-MM
	Profit   float64 //獲利
}

type CustomerReportSummaryUI struct {
	CustomerReportSummary
	StartDate string
	EndDate   string
}
