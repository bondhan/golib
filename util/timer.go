package util

import (
	"fmt"
	"time"
)

// Timer simple timer
type Timer struct {
	timeout time.Duration
	start   time.Time
}

// NewTimer create timer instance
func NewTimer(to int) *Timer {
	return &Timer{
		timeout: time.Duration(to) * time.Second,
		start:   time.Now(),
	}
}

// IsTimeout is timeout already
func (t *Timer) IsTimeout() bool {
	return time.Since(t.start) > t.timeout
}

// Extend extend duration
func (t *Timer) Extend(to int) {
	t.timeout += time.Duration(to) * time.Second
}

func NormalizeTime(t int64, out string) int64 {
	ts := getTime(t)

	switch out {
	case "ms":
		return ts.UnixMilli()
	case "us":
		return ts.UnixMicro()
	case "ns":
		return ts.UnixNano()
	default:
		return ts.Unix()
	}
}

func GetNormalizedTime(t int64) time.Time {
	return getTime(t)
}

func getTime(t int64) time.Time {
	tstring := fmt.Sprintf("%v", t)

	//Nano second
	if len(tstring) > 18 {
		return time.Unix(t/1000000000, t%1000000000)
	}
	//Micro second
	if len(tstring) > 15 && len(tstring) <= 18 {
		return time.UnixMicro(t)
	}

	//Mili second
	if len(tstring) > 12 && len(tstring) <= 15 {
		return time.UnixMilli(t)
	}

	return time.Unix(t, 0)
}

func SecondsTillMidnight(tz string) int {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc, _ = time.LoadLocation("")
	}
	y, m, d := time.Now().Date()
	midnight := time.Date(y, m, d, 23, 59, 59, 0, loc)
	return int(time.Until(midnight).Seconds())
}
