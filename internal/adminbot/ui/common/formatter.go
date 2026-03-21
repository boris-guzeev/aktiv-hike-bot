package common

import "time"

func Format(t time.Time) string {
	return t.Format("02 Jan 2006 15:04")
}
