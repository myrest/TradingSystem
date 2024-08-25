package common

import (
	"reflect"
	"testing"
	"time"
)

func TestGetWeeksInDateRange(t *testing.T) {
	type args struct {
		startDate time.Time
		endDate   time.Time
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Single week in same year",
			args: args{
				startDate: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
				endDate:   time.Date(2024, time.January, 7, 0, 0, 0, 0, time.UTC),
			},
			want: []string{"2024-01"}, // 2024年第1週
		},
		{
			name: "Two weeks in same year",
			args: args{
				startDate: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
				endDate:   time.Date(2024, time.January, 14, 0, 0, 0, 0, time.UTC),
			},
			want: []string{"2024-01", "2024-02"}, // 包含第2週
		},
		{
			name: "Cross year from December to January",
			args: args{
				startDate: time.Date(2024, time.December, 28, 0, 0, 0, 0, time.UTC),
				endDate:   time.Date(2025, time.January, 4, 0, 0, 0, 0, time.UTC),
			},
			want: []string{"2024-52", "2025-01"}, // 包含第52週和第1週
		},
		{
			name: "Full year span",
			args: args{
				startDate: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
				endDate:   time.Date(2024, time.December, 31, 0, 0, 0, 0, time.UTC),
			},
			want: []string{ // 包含所有週
				"2024-01", "2024-02", "2024-03", "2024-04", "2024-05", "2024-06",
				"2024-07", "2024-08", "2024-09", "2024-10", "2024-11", "2024-12",
				"2024-13", "2024-14", "2024-15", "2024-16", "2024-17", "2024-18",
				"2024-19", "2024-20", "2024-21", "2024-22", "2024-23", "2024-24",
				"2024-25", "2024-26", "2024-27", "2024-28", "2024-29", "2024-30",
				"2024-31", "2024-32", "2024-33", "2024-34", "2024-35", "2024-36",
				"2024-37", "2024-38", "2024-39", "2024-40", "2024-41", "2024-42",
				"2024-43", "2024-44", "2024-45", "2024-46", "2024-47", "2024-48",
				"2024-49", "2024-50", "2024-51", "2024-52", "2025-01",
			},
		},
		{
			name: "Start and end the same day",
			args: args{
				startDate: time.Date(2024, time.February, 1, 0, 0, 0, 0, time.UTC),
				endDate:   time.Date(2024, time.February, 1, 0, 0, 0, 0, time.UTC),
			},
			want: []string{"2024-05"}, // 只有第5週
		},
		{
			name: "End date earlier than start date",
			args: args{
				startDate: time.Date(2024, time.February, 15, 0, 0, 0, 0, time.UTC),
				endDate:   time.Date(2024, time.February, 1, 0, 0, 0, 0, time.UTC),
			},
			want: nil, // 應該返回空
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetWeeksInDateRange(tt.args.startDate, tt.args.endDate); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetWeeksInDateRange() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWeekToDateRange(t *testing.T) {
	type args struct {
		weekStr string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   string
		wantErr bool
	}{
		{
			name:    "Valid Case - Week 1 of 2025",
			args:    args{weekStr: "2025-01"},
			want:    "2024-12-30", // 2025年第1週的起始日 (週一)
			want1:   "2025-01-05", // 2025年第1週的結束日 (週日)
			wantErr: false,
		},
		{
			name:    "Valid Case - Week 1 of 2025",
			args:    args{weekStr: "2025-02"},
			want:    "2025-01-06", // 2025年第1週的起始日 (週一)
			want1:   "2025-01-12", // 2025年第1週的結束日 (週日)
			wantErr: false,
		},
		{
			name:    "Valid Case - Week 52 of 2025",
			args:    args{weekStr: "2024-52"},
			want:    "2024-12-23", // 2025年第52週的起始日 (週一)
			want1:   "2024-12-29", // 2025年第52週的結束日 (週日)
			wantErr: false,
		},
		{
			name:    "Valid Case - Week 53 of 2025",
			args:    args{weekStr: "2024-53"},
			want:    "2024-12-30", // 2025年第53週的起始日 (週一)
			want1:   "2025-01-05", // 2025年第53週的結束日 (週日)
			wantErr: false,
		},
		{
			name:    "Valid Case - Week 1 of 2024",
			args:    args{weekStr: "2024-01"},
			want:    "2024-01-01", // 2024年第1週的起始日 (週一)
			want1:   "2024-01-07", // 2024年第1週的結束日 (週日)
			wantErr: false,
		},
		{
			name:    "Invalid Case - Invalid Format",
			args:    args{weekStr: "2024-99"},
			want:    "", // 期待的結果不應該有
			want1:   "",
			wantErr: true,
		},
		{
			name:    "Invalid Case - Negative Year",
			args:    args{weekStr: "-2024-01"},
			want:    "", // 期待的結果不應該有
			want1:   "",
			wantErr: true,
		},
		{
			name:    "Invalid Case - Week 0",
			args:    args{weekStr: "2024-00"},
			want:    "", // 期待的結果不應該有
			want1:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := WeekToDateRange(tt.args.weekStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("WeekToDateRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("WeekToDateRange() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("WeekToDateRange() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestGetWeeksByDate(t *testing.T) {
	type args struct {
		date time.Time
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetWeeksByDate(tt.args.date); got != tt.want {
				t.Errorf("GetWeeksByDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetUtcTimeNow(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetUtcTimeNow(); got != tt.want {
				t.Errorf("GetUtcTimeNow() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatTime(t *testing.T) {
	type args struct {
		t time.Time
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatTime(tt.args.t); got != tt.want {
				t.Errorf("FormatTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseTime(t *testing.T) {
	type args struct {
		timeStr string
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseTime(tt.args.timeStr); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetPreviousMondays(t *testing.T) {
	type args struct {
		date time.Time
		n    int
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Case 1: 5 weeks back from a Monday",
			args: args{
				date: time.Date(2024, 8, 26, 0, 0, 0, 0, time.UTC), // 一個星期一
				n:    5,
			},
			want: []string{
				"2024-08-26", // 第1個星期一
				"2024-08-19", // 第2個星期一
				"2024-08-12", // 第3個星期一
				"2024-08-05", // 第4個星期一
				"2024-07-29", // 第5個星期一
			},
		},
		{
			name: "Case 2: 5 weeks back from a non-Monday",
			args: args{
				date: time.Date(2024, 8, 30, 0, 0, 0, 0, time.UTC), // 一個星期五
				n:    5,
			},
			want: []string{
				"2024-08-26", // 第1個星期一
				"2024-08-19", // 第2個星期一
				"2024-08-12", // 第3個星期一
				"2024-08-05", // 第4個星期一
				"2024-07-29", // 第5個星期一
			},
		},
		{
			name: "Case 3: N is zero",
			args: args{
				date: time.Date(2024, 8, 30, 0, 0, 0, 0, time.UTC),
				n:    0,
			},
			want: []string{}, // 沒有星期一
		},
		{
			name: "Case 4: N is negative",
			args: args{
				date: time.Date(2024, 8, 30, 0, 0, 0, 0, time.UTC),
				n:    -2,
			},
			want: []string{}, // 沒有星期一
		},
		{
			name: "Case 5: Edge case - New Year",
			args: args{
				date: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), // 新年
				n:    3,
			},
			want: []string{
				"2024-01-01", // 第1個星期一
				"2023-12-25", // 第2個星期一
				"2023-12-18", // 第3個星期一
			},
		},
		{
			name: "Case 6: Edge case - Leap Year",
			args: args{
				date: time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC), // 閏年
				n:    4,
			},
			want: []string{
				"2024-02-26", // 第1個星期一
				"2024-02-19", // 第2個星期一
				"2024-02-12", // 第3個星期一
				"2024-02-05", // 第4個星期一
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetPreviousMondays(tt.args.date, tt.args.n)
			if err != nil {
				t.Errorf("GetPreviousMondays() error = %v", err)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("GetPreviousMondays() got = %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("GetPreviousMondays() got[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}
