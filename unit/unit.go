package unit

import (
	"fmt"
	"time"
)

func FormatBytes(n uint64) (result string) {
	switch {
	case n >= 1024*1024*1024:
		result = fmt.Sprintf("%.02f GB", float64(n)/(1024*1024*1024))
	case n >= 1024*1024:
		result = fmt.Sprintf("%.02f MB", float64(n)/(1024*1024))
	case n >= 1024:
		result = fmt.Sprintf("%.02f KB", float64(n)/1024)
	default:
		result = fmt.Sprintf("%d B", n)
	}

	return
}

func FormatDuration(d time.Duration) (result string) {
	if d > time.Hour*24 {
		result = fmt.Sprintf("%dd", d/24/time.Hour)
		d -= (d / time.Hour / 24) * (time.Hour * 24)
	}

	if d > time.Hour {
		result = fmt.Sprintf("%s%dh", result, d/time.Hour)
		d -= d / time.Hour * time.Hour
	}

	if d > time.Minute {
		result = fmt.Sprintf("%s%dm", result, d/time.Minute)
		d -= d / time.Minute * time.Minute
	}

	s := d / time.Second

	result = fmt.Sprintf("%s%ds", result, s)

	return
}
