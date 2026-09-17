package sportslot

import (
	"net/http"
	"time"
)

type TimeRange struct {
	From time.Time
	To   time.Time
}

func WeekRangeAmericaLima(now time.Time) TimeRange {
	loc := time.FixedZone("America/Lima", -5*60*60)
	d := now.In(loc)
	y, m, day := d.Date()
	// weekday: Sunday=0
	wd := int(d.Weekday())
	// make Monday=0
	// Monday(1)->0, Sunday(0)->6
	diff := (wd + 6) % 7
	start := time.Date(y, m, day, 0, 0, 0, 0, loc).AddDate(0, 0, -diff)
	end := start.AddDate(0, 0, 7)
	return TimeRange{From: start, To: end}
}

func ParseTimeInLima(s string) (time.Time, error) {
	loc := time.FixedZone("America/Lima", -5*60*60)
	// Accept RFC3339; if no zone, parse as local Lima.
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.In(loc), nil
	}
	if t, err := time.ParseInLocation("2006-01-02T15:04:05", s, loc); err == nil {
		return t, nil
	}
	return time.Time{}, http.ErrBodyNotAllowed
}
