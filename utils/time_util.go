package utils

import "time"

func GetCurrentDatetime() string {
	now := time.Now()

	return now.Format("2006-01-02 15:04:05")
}
