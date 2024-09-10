package common

import (
	"fmt"
	"time"
)

const DateTimeLayout = "2006-01-02 15:04:05"

// GetWeeksByDate 傳入一個日期回傳該年份的第幾週
func GetWeeksByDate(date time.Time) string {
	weeks := GetWeeksInDateRange(date, date)
	if len(weeks) > 0 {
		return weeks[0]
	} else {
		return ""
	}
}

func GetWeeksStartEndDateByDate(date time.Time) (string, string) {
	week := GetWeeksByDate(date)
	sd, ed, _ := WeekToDateRange(week)
	return sd, ed
}

// WeekToDateRange 根據給定的 YYYY-WW 格式計算起始和結束日期
func WeekToDateRange(weekStr string) (string, string, error) {
	var year, week int
	_, err := fmt.Sscanf(weekStr, "%d-%d", &year, &week)
	if year < 0 || week < 1 {
		return "", "", fmt.Errorf("invalid week number: %d", week)
	}
	if err != nil {
		return "", "", err
	}

	if week < 1 || week > 53 {
		return "", "", fmt.Errorf("invalid week number: %d", week)
	}

	// 找到該年的第一天
	startOfYear := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)

	// 找到該年的第一個週一
	for startOfYear.Weekday() != time.Monday {
		startOfYear = startOfYear.AddDate(0, 0, -1)
	}

	// 計算該週的起始日期（週一）
	startDate := startOfYear.AddDate(0, 0, (week-1)*7)

	// 計算該週的結束日期（週日）
	endDate := startDate.AddDate(0, 0, 6)

	return startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), nil
}

func GetUtcTimeNow() string {
	return time.Now().UTC().Format(DateTimeLayout)
}

func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// 回傳日期區間的週數
func GetWeeksInDateRange(startDate, endDate time.Time) []string {
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

// ParseTime 將字串轉換為 time.Time，支持日期和日期時間格式
func ParseTime(timeStr string) time.Time {
	// 定義可能的時間格式
	formats := []string{
		"2006-01-02 15:04:05", // 日期時間格式
		"2006-01-02",          // 僅日期格式
	}

	// 嘗試使用所有格式解析
	for _, layout := range formats {
		if t, err := time.Parse(layout, timeStr); err == nil {
			return t // 成功解析，退出循環
		}
	}

	return TimeMax()
}

func TimeMax() time.Time {
	return time.Time{}.AddDate(9999, 12, 31)
}

// GetPreviousMondays 根據傳入的日期，返回前 N 個星期一的日期
func GetPreviousMondays(date time.Time, n int) ([]string, error) {
	var mondays []string

	// 找到最近的星期一
	offset := int(date.Weekday())                 // 0=Sunday, 1=Monday, ..., 6=Saturday
	offset = offset - 1                           //如果是星期一，就減掉0天，如果是星期二就要減掉一天
	previousMonday := date.AddDate(0, 0, -offset) // 計算最近的星期一

	// 加入前 N 個星期一
	for i := 0; i < n; i++ {
		mondays = append(mondays, previousMonday.Format("2006-01-02"))
		previousMonday = previousMonday.AddDate(0, 0, -7) // 向前推算 7 天
	}

	return mondays, nil
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
