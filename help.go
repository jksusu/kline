package kline

// 时间取模算法
func GetTimeModeling(period string, t int64) int64 {
	var timestamp int64
	switch period {
	case AMinute:
		timestamp = t - (t % int64(60))
		break
	case FiveMinutes:
		timestamp = t - (t % int64(5*60))
		break
	case FifteenMinutes:
		timestamp = t - (t % int64(15*60))
		break
	case Minutes:
		timestamp = t - (t % int64(30*60))
		break
	case AnHour:
		timestamp = t - (t % int64(60*60))
		break
	case FourHours:
		timestamp = t - (t % int64(4*60*60))
		break
	case ADay:
		timestamp = t - (t % int64(24*60*60))
		break
	case AWeek:
		timestamp = t - (t % int64(7*24*60*60))
		break
	case OneMonth:
		timestamp = t - (t % int64(30*24*60*60))
		break
	case AYear:
		timestamp = t - (t % int64(365*24*60*60))
		break
	default:
		return t
	}
	return timestamp
}
