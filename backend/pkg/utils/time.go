package utils

import (
	"time"
	_ "time/tzdata"
)

var (
	// China Standard Time UT+8:00
	cstLocation *time.Location
)

func init() {
	var err error
	if cstLocation, err = time.LoadLocation("Asia/Shanghai"); err != nil {
		panic(err)
	}
}

// TimeParse 按RFC3339标准解析时间
func TimeParse(value string) (time.Time, error) {
	return time.Parse(time.RFC3339, value)
}

// TimeFormat 时间Format为中国时间字符串
func TimeFormat(t time.Time) string {
	return t.In(cstLocation).Format("2006-01-02 15:04:05")
}

func Sec2Duration(sec int64) time.Duration {
	return time.Duration(sec) * time.Second
}

func Duration2Sec(dur time.Duration) int64 {
	return int64(dur.Seconds())
}
