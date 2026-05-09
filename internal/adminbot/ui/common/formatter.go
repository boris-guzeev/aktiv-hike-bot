package common

import "time"

func Format(t time.Time) string {
	return t.Format("02.01.2006")
}
