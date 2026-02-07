package utils

import "time"

const timeLayout = "2006-01-02T15:04:05"

func ParseLocalAt(localISO string, loc *time.Location) (time.Time, error) {
	return time.ParseInLocation(timeLayout, localISO, loc)
}

func Minutes(d time.Duration) int {
	if d < 0 {
		return int((d - time.Second) / time.Minute)
	}
	return int(d / time.Minute)
}
