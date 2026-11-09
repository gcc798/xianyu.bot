package security

import (
	"strconv"
	"strings"
	"time"
)

// ParseDuration 解析 Go duration、带 d 的天数或纯纳秒整数。
func ParseDuration(d string) (time.Duration, error) {
	d = strings.TrimSpace(d)
	dr, err := time.ParseDuration(d)
	if err == nil {
		return dr, nil
	}
	if daysText, remainder, found := strings.Cut(d, "d"); found {
		days, err := strconv.Atoi(daysText)
		if err != nil {
			return 0, err
		}
		daysDuration := 24 * time.Hour * time.Duration(days)
		if remainder == "" {
			return daysDuration, nil
		}
		remainderDuration, err := time.ParseDuration(remainder)
		if err != nil {
			return 0, err
		}
		return daysDuration + remainderDuration, nil
	}

	dv, err := strconv.ParseInt(d, 10, 64)
	return time.Duration(dv), err
}
