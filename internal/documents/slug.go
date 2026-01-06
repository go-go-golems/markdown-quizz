package documents

import (
	"regexp"
	"strings"
	"time"
)

var slugTitleRegexp = regexp.MustCompile(`[^a-z0-9]+`)

func GenerateSlug(title string, now time.Time) string {
	base := strings.ToLower(title)
	base = slugTitleRegexp.ReplaceAllString(base, "-")
	base = strings.Trim(base, "-")
	return base + "-" + strconvFormatInt36(now.UnixMilli())
}

func strconvFormatInt36(v int64) string {
	const digits = "0123456789abcdefghijklmnopqrstuvwxyz"
	if v == 0 {
		return "0"
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var buf [32]byte
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = digits[int(v%36)]
		v /= 36
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
