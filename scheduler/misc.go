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
