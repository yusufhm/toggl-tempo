package time

import (
	"time"
)

func WeekStartDate(date time.Time) time.Time {
	offset := (int(time.Monday) - int(date.Weekday()) - 7) % 7
	startDateTime := date.Add(time.Duration(offset*24) * time.Hour)
	result := time.Date(startDateTime.Year(), startDateTime.Month(), startDateTime.Day(), 0, 0, 0, 0, time.Local)
	return result
}
