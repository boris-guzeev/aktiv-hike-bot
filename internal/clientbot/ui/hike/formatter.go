package hike

import (
	"fmt"
	"time"
)

func FormatDateRange(start, end time.Time) string {
	startDate := start.Format("02.01.2006")
	endDate := end.Format("02.01.2006")

	if startDate == endDate {
		return startDate
	}

	return fmt.Sprintf("%s — %s", startDate, endDate)
}
