package common

import (
	"time"
)

const DateTimeLayout = "2006-01-02 15:04:05"

func GetUtcTimeNow() string {
	return time.Now().UTC().Format(DateTimeLayout)
}

func FormatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// ParseTime 將字串轉換為 time.Time，支持日期和日期時間格式
func ParseTime(timeStr string) time.Time {
	// 定義可能的時間格式
	formats := []string{
		"2006-01-02 15:04:05", // 日期時間格式-秒
		"2006-01-02 15:04",    // 日期時間格式-分
		"2006-01-02",          // 僅日期格式
		"2006-01",             // 僅月份格式
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
