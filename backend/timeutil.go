package main

import "time"

const timeLayout = "2006-01-02T15:04:05"

func ParseLocalAt(localISO string, loc *time.Location) (time.Time, error) {
	return time.ParseInLocation(timeLayout, localISO, loc)
}

func DepartureUTC(s *Store, f Flight) (time.Time, error) {
	loc, err := s.TZLocation(f.Origin)
	if err != nil {
		return time.Time{}, err
	}
	t, err := ParseLocalAt(f.DepartureTime, loc)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}

func ArrivalUTC(s *Store, f Flight) (time.Time, error) {
	loc, err := s.TZLocation(f.Destination)
	if err != nil {
		return time.Time{}, err
	}
	t, err := ParseLocalAt(f.ArrivalTime, loc)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}

func Minutes(d time.Duration) int {
	if d < 0 {
		return int((d - time.Second) / time.Minute)
	}
	return int(d / time.Minute)
}
