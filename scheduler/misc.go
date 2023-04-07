package scheduler

import "time"

func parseTime(value string, layout []string) (t time.Time, err error) {
	for _, layout := range layout {
		t, err = time.Parse(layout, value)
		if err == nil {
			return
		}
	}
	return
}

func gcd(a, b time.Duration) time.Duration {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func first(t time.Time, d time.Duration) time.Duration {
	return t.Add(d).Truncate(d).Sub(t)
}
