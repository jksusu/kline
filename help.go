package kline

// 时间取模算法
func GetTimeModeling(period string, t int64) int64 {
	// 定义时间段映射
	periodMap := map[string]int64{
		AMinute:        60,
		FiveMinutes:    5 * 60,
		FifteenMinutes: 15 * 60,
		Minutes:        30 * 60,
		AnHour:         60 * 60,
		TwoHours:       2 * 60 * 60,
		FourHours:      4 * 60 * 60,
		ADay:           24 * 60 * 60,
		AWeek:          7 * 24 * 60 * 60,
		OneMonth:       30 * 24 * 60 * 60,
		AYear:          365 * 24 * 60 * 60,
	}
	if periodValue, exists := periodMap[period]; exists {
		return t - (t % periodValue)
	}
	return t
}
