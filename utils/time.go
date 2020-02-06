package utils

import (
	"time"
)

func Now() time.Time {
	return time.Now().UTC()
}

func DayNow() time.Time {
	tNow := Now()
	return time.Date(tNow.Year(), tNow.Month(), tNow.Day(), 0, 0, 0, 0, time.UTC)
}

func MonthNow() time.Time {
	tNow := Now()
	return time.Date(tNow.Year(), tNow.Month(), 1, 0, 0, 0, 0, time.UTC)
}
