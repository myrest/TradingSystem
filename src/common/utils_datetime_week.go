package common

import (
	"fmt"
	"time"
)

// GetWeeksByDate 傳入一個日期回傳該年份的第幾週
func GetWeeksByDate(date time.Time) string {
	weeks := GetWeeksInDateRange(date, date)
	if len(weeks) > 0 {
		return weeks[0]
	} else {
		return ""
	}
}

func GetWeeksStartEndDateByDate(date time.Time) (time.Time, time.Time) {
	week := GetWeeksByDate(date)
	sd, ed, _ := WeekToDateRange(week)
	return sd, ed
}

// WeekToDateRange 根據給定的 YYYY-WW 格式計算起始和結束日期
func WeekToDateRange(weekStr string) (time.Time, time.Time, error) {
	var year, week int
	_, err := fmt.Sscanf(weekStr, "%d-%d", &year, &week)
	if year < 0 || week < 1 {
		return time.Now().UTC(), time.Now().UTC(), fmt.Errorf("invalid week number: %d", week)
	}
	if err != nil {
		return time.Now().UTC(), time.Now().UTC(), err
	}

	if week < 1 || week > 53 {
		return time.Now().UTC(), time.Now().UTC(), fmt.Errorf("invalid week number: %d", week)
	}

	// 找到該年的第一天
	startOfYear := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)

	// 找到該年的第一個週一
	for startOfYear.Weekday() != time.Monday {
		startOfYear = startOfYear.AddDate(0, 0, -1)
	}

	// 計算該週的起始日期（週一）
	startDate := startOfYear.AddDate(0, 0, (week-1)*7).Truncate(24 * time.Hour) //去時分秒

	// 計算該週的結束日期（週日）加上時分秒
	endDate := startDate.AddDate(0, 0, 6)
	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, endDate.Location())

	return startDate, endDate, nil
}

// 回傳日期區間的週數
func GetWeeksInDateRange(startDate, endDate time.Time) []string {
	//要依星期一，為一週的開始，計算才會正確。
	mondays, _ := GetPreviousMondays(startDate, 1)
	startDate = mondays[0]

	if endDate.Sub(startDate) > 365*24*time.Hour {
		return nil
	}

	var weeks []string
	current := startDate

	for (current.After(startDate) || current.Equal(startDate)) && (current.Before(endDate) || current.Equal(endDate)) {
		year, week := current.ISOWeek()
		weekStr := fmt.Sprintf("%d-%02d", year, week)

		// 檢查當前週數是否是同一日期區間的末尾週數
		if len(weeks) > 0 {
			lastWeekYear, lastWeekNum := current.AddDate(0, 0, +7).ISOWeek()
			if year == lastWeekYear && week == lastWeekNum+1 {
				// 如果當前週數是下一週的上週，則跳過
				current = current.AddDate(0, 0, 7)
				continue
			}
		}

		weeks = append(weeks, weekStr)
		current = current.AddDate(0, 0, 7)
	}

	return weeks
}

// GetPreviousMondays 根據傳入的日期，返回前 N 個星期一的日期
func GetPreviousMondays(date time.Time, n int) ([]time.Time, error) {
	date = date.Truncate(24 * time.Hour) //去掉時分秒
	var mondays []time.Time

	// 找到最近的星期一
	offset := int(date.Weekday())                 // 0=Sunday, 1=Monday, ..., 6=Saturday
	offset = offset - 1                           //如果是星期一，就減掉0天，如果是星期二就要減掉一天
	previousMonday := date.AddDate(0, 0, -offset) // 計算最近的星期一

	// 加入前 N 個星期一
	for i := 0; i < n; i++ {
		mondays = append(mondays, previousMonday)
		previousMonday = previousMonday.AddDate(0, 0, -7) // 向前推算 7 天
	}

	return mondays, nil
}
