package libs

import (
	"strconv"
	"time"
)

const (
	DateTimeSecFormatter    = "2006-01-02 15:04:05"
	DayTimeMilliFormatter   = "2006-01-02 15:04:05.000"
	DateTimeMinuteFormatter = "2006-01-02 15:04"
	DateFormatter           = "2006-01-02"
	DateDigitFormatter      = "20060102"
	TimeFormatter           = "15:04:05"
	DateTimeCronFormatter   = "05 04 15 02 01 *"
)

const (
	Second       = 1
	Minute       = 60
	HalfHour     = 1800
	Hour         = 3600
	HalfDay      = 43200
	HalfDayMilli = 43200000
	Day          = 86400
)

func StartOfDay(t time.Time) int64 {
	y, m, d := t.Date()
	t1 := time.Date(y, m, d, 0, 0, 0, 0, t.Location())
	return t1.Unix()
}

func EndOfDay(t time.Time) int64 {
	y, m, d := t.Date()
	t1 := time.Date(y, m, d+1, 0, 0, 0, 0, t.Location())
	return t1.Unix() - 1
}

func UTC() time.Time {
	return time.Now().UTC()
}

func StartOfDayLoc(t time.Time, location *time.Location) int64 {
	t1 := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, location)
	return t1.Unix()
}

func UnixSecs() int64 {
	return time.Now().Unix()
}

func UnixMilli() int64 {
	return time.Now().UnixMilli()
}

func UnixMicro() int64 {
	return time.Now().UnixMicro()
}

func UnixNano() int64 {
	return time.Now().UTC().UnixNano()
}

func DayStartWithLoc(t time.Time, loc *time.Location) (time.Time, error) {
	return time.ParseInLocation(DateFormatter, t.Format(DateFormatter), loc)
}

func DayEndWithLoc(t time.Time, loc *time.Location) (time.Time, error) {
	start, err := DayStartWithLoc(t, loc)
	if err != nil {
		return start, err
	}
	return start.Add(time.Hour*24 - time.Second), nil
}

func ToTimestamp(datetime string) int64 {
	t, _ := time.ParseInLocation(DateTimeSecFormatter, datetime, time.Local)
	return t.Unix()
}

func TimeFormat(timestamp int64, format string) string {
	if format != "" {
		return time.Unix(timestamp, 0).Format(format)
	} else {
		return time.Unix(timestamp, 0).Format(DateTimeSecFormatter)
	}
}

func RfcToTime(datetime string, isUTC bool) time.Time {
	t, _ := time.Parse(time.RFC3339, datetime)
	if !isUTC {
		return t
	} else {
		return t.UTC()
	}
}

func YmdInt(t time.Time) int64 {
	n, _ := strconv.ParseInt(t.Format(DateDigitFormatter), 10, 64)
	return n
}

func DateDiff(t1, t2, format string, loc *time.Location) int64 {
	time1, _ := time.ParseInLocation(format, t1, loc)
	time2, _ := time.ParseInLocation(format, t2, loc)
	if time1.Unix() > time2.Unix() {
		return (time1.Unix() - time2.Unix()) / (24 * 60 * 60)
	} else {
		return (time2.Unix() - time1.Unix()) / (24 * 60 * 60)
	}
}

func IsTime(value, format string) bool {
	_, err := time.Parse(format, value)
	return err == nil
}
