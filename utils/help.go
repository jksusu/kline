package utils

import (
	"strconv"
)

func ConvertStringToFloat64(str string) float64 {
	value, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0
	}
	return value
}

func ConvertStringToInt(str string) int {
	num, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return num
}

func ConvertStringToInt64(str string) int64 {
	num, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0
	}
	return num
}
