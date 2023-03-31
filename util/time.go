package util

import (
	"encoding/base64"
	"strconv"
	"strings"
	"time"
)

const JakartaWibTz = "Asia/Jakarta"
const tzData = "VFppZjIAAAAAAAAAAAAAAAAAAAAAAAAGAAAABgAAAAAAAAAHAAAABgAAABypeIXguhbeYMu/g4jSVu5w1zzGCNr/JgD0tb6IAQIDAgQCBQAAZCAAAAAAZyAABAAAaXgACgAAfpAAEAAAcIAAFAAAYnAAGEJNVAArMDcyMAArMDczMAArMDkAKzA4AFdJQgAAAAAAAAAAAAAAAABUWmlmMgAAAAAAAAAAAAAAAAAAAAAAAAcAAAAHAAAAAAAAAAgAAAAHAAAAIP////8/Zklg/////6l4heD/////uhbeYP/////Lv4OI/////9JW7nD/////1zzGCP/////a/yYA//////S1vogBAgMEAwUDBgAAZCAAAAAAZCAABAAAZyAACAAAaXgADgAAfpAAFAAAcIAAGAAAYnAAHExNVABCTVQAKzA3MjAAKzA3MzAAKzA5ACswOABXSUIAAAAAAAAAAAAAAAAAAAAKV0lCLTcK"
const MongoDateFormat = "2006-01-02T15:04:05.000-07:00"

func GetTimeInLocation(t time.Time, name string) (timeLocation time.Time, err error) {
	tb, err := base64.RawStdEncoding.DecodeString(tzData)
	if err != nil {
		return
	}
	loc, err := time.LoadLocationFromTZData(name, tb)
	if err == nil {
		timeLocation = t.In(loc)
	}
	return
}

func GetTimeInJakartaLocation(t time.Time) (timeLocation time.Time, err error) {
	return GetTimeInLocation(t, JakartaWibTz)
}

func DateStringToTime(v string) (timeLocation time.Time, err error) {
	if vi, err := strconv.ParseInt(v, 10, 64); err == nil {
		return getTime(vi), nil
	}
	switch len(v) {
	case len(DateFormat):
		return time.Parse(DateFormat, v)
	case len(DateTimeFormat):
		if strings.Contains(v, "T") {
			return time.Parse(DateTimeFormat, v)
		}
		return time.Parse(DateTimeSpaceFormat, v)
	case len(MongoDateFormat):
		return time.Parse(MongoDateFormat, v)
	default:
		return time.Parse(time.RFC3339, v)
	}
}

func IsSameDate(source int64, target int64) bool {
	dataSource := time.Unix(source, 0).Format("2006-01-02")
	dataTarget := time.Unix(target, 0).Format("2006-01-02")
	return dataSource == dataTarget
}
