package util

import (
	"strconv"
	"time"
)

const (
	TZLocStr         = "Asia/Taipei"
	FullDateLayout   = "2006-01-02T15:04:05.999999-07:00"
	SimpleDateLayout = "2006-01-02 15:04:05"
	DateLayout       = "2006-01-02"
)

var serviceTimeLoc *time.Location

func init() {
	serviceTimeLoc = time.FixedZone("GMT+8", 8*60*60)
}

func GetServiceTimeLoc() *time.Location {
	return serviceTimeLoc
}

func GetFullDateStr(t time.Time) string {
	return t.In(GetServiceTimeLoc()).Format(FullDateLayout)
}

func GetSimpleDateStr(t time.Time) string {
	return t.In(GetServiceTimeLoc()).Format(SimpleDateLayout)
}

func GetDateStr(t time.Time) string {
	return t.In(GetServiceTimeLoc()).Format(DateLayout)
}

func GetTimeValue(t time.Time) (string, int64) {
	return GetFullDateStr(t), GetTimestampMS(t)
}

func GetTimeValuePtr(t *time.Time) (*string, *int64) {
	if t == nil {
		return nil, nil
	}
	str := GetFullDateStr(*t)
	unix := GetTimestampMS(*t)
	return &str, &unix
}

func GetTimestampMS(t time.Time) int64 {
	return t.Unix() * 1000
}

func CalculateEndTimeOfDay(t time.Time) time.Time {
	localTime := t.In(serviceTimeLoc)
	return time.Date(localTime.Year(), localTime.Month(), localTime.Day(), 23, 59, 59, 0, localTime.Location())
}

func MsTimestampStrToTime(str string) *time.Time {
	// ignore ms part
	sTimestamp, err := strconv.Atoi(str[:len(str)-3])
	if err != nil {
		return nil
	}
	t := time.Unix(int64(sTimestamp), 0)

	return &t
}
