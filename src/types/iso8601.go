package types

import (
	"errors"
	"time"

	"github.com/kodergarten/iso8601duration"
)

type ISO8601Date = string
type ISO8601Duration = string

func ParseISO8601Duration(val ISO8601Duration, minDuration time.Duration) (dur time.Duration, err error) {
	duration, err := iso8601duration.ParseString(val)
	if err != nil {
		return
	}
	dur = duration.ToDuration()
	if dur < minDuration {
		err = errors.New("field Duration needs to be higher")
		return
	}
	return
}
