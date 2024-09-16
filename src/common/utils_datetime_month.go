package common

import "time"

// 回傳從今天起算，前三個月的1號日期
func GetMonthlyDay1(count int) []time.Time {
	var result []time.Time
	for i := 0; i < count; i++ {
		firstDay := time.Now().AddDate(0, -i, 0).AddDate(0, 0, -time.Now().Day()+1)
		result = append(result, firstDay)
	}
	return result
}

// 依日期，回傳該月份起迄日期
func GetMonthStartEndDate(date time.Time) (time.Time, time.Time) {
	// 取得該日期的年份和月份
	year, month, _ := date.Date()
	// 取得該月份的第一天
	start := time.Date(year, month, 1, 0, 0, 0, 0, date.Location())
	// 取得該月份的最後一天
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond) // 加一個月再減去一個納秒
	return start, end
}

// GetMonthsInRange 接收起始和結束時間，返回中間所有月份的資料陣列，格式為 YYYY-MM
func GetMonthsInRange(dt ...time.Time) []string {
	start := dt[0]
	end := dt[0]
	if len(dt) > 1 {
		end = dt[1]
	}
	var months []string
	// 取得起始月份
	current := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, start.Location())
	for current.Before(end) || current.Equal(end) {
		months = append(months, current.Format("2006-01"))
		// 移動到下一個月份
		current = current.AddDate(0, 1, 0)
	}
	return months
}
