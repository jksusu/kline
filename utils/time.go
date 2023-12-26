package utils

import "time"

// 时间转时间戳
func ParseDate(dateStr string, layout string) int64 {
	t, err := time.Parse(layout, dateStr)
	if err != nil {
		return 0
	}
	return t.Unix()
}
