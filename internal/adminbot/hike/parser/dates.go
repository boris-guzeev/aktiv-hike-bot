package parser

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	reDay      = regexp.MustCompile(`^\d{1,2}$`)             // 1..31
	reDayMonth = regexp.MustCompile(`(^\d{1,2})\.(\d{1,2})`) // dd.mm
)

// TODO: Вынести парсеры в отдельный файл
func ParseHikeDates(input string, now time.Time, loc *time.Location) (time.Time, time.Time, error) {
	s := strings.TrimSpace(input)
	s = strings.ReplaceAll(s, ",", " ")
	s = strings.ReplaceAll(s, "—", "-")
	s = strings.ReplaceAll(s, "–", "-")
	s = strings.ReplaceAll(s, "  ", " ")

	// Normalize range delimimiters
	// cases: "10 12", "10-12", "10 - 12", "03.02 04.02", "03.02-04.02"
	rangeDelim := regexp.MustCompile(`\s*-\s*|\s+`)
	tokens := rangeDelim.Split(s, -1)

	switch len(tokens) {
	case 1:
		// One value: either "10" or "03.02"
		return parseSingle(tokens[0], now, loc)

	case 2:
		// Range: "10 12", "31 3", "03.02 04.02", "15.12 16.12", "03.02-04.02"
		return parseRange(tokens[0], tokens[1], now, loc)

	default:
		return time.Time{}, time.Time{}, errors.New("не получилось распознать даты (слишком много частей)")
	}
}

func parseSingle(token string, now time.Time, loc *time.Location) (time.Time, time.Time, error) {
	if reDayMonth.MatchString(token) {
		start, err := parseDDMM(token, now.Year(), loc)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		today := truncateToDay(now.In(loc))
		if start.Before(today) {
			start, err = parseDDMM(token, now.Year()+1, loc)
			if err != nil {
				return time.Time{}, time.Time{}, err
			}
		}
		// один день
		end := time.Date(start.Year(), start.Month(), start.Day(), 22, 0, 0, 0, loc)
		return start, end, nil
	}

	if reDay.MatchString(token) {
		day, _ := strconv.Atoi(token)

		var start time.Time
		// выбираем месяц: если день уже прошёл — берём следующий
		if now.Day() > day {
			start = time.Date(now.Year(), now.Month()+1, day, 8, 0, 0, 0, loc)
		} else {
			start = time.Date(now.Year(), now.Month(), day, 8, 0, 0, 0, loc)
		}

		// однодневный хайк
		end := time.Date(start.Year(), start.Month(), start.Day(), 22, 0, 0, 0, loc)
		return start, end, nil
	}

	return time.Time{}, time.Time{}, errors.New("ожидал формат 'dd' или 'dd.mm'")
}

func parseRange(a, b string, now time.Time, loc *time.Location) (time.Time, time.Time, error) {
	switch {
	case reDay.MatchString(a) && reDay.MatchString(b):
		// both are days without months
		return parseDayDay(a, b, now, loc)
	case reDayMonth.MatchString(a) && reDayMonth.MatchString(b):
		// both are dd.mm
		return parseDDMM_DDMM(a, b, now, loc)
	default:
		return time.Time{}, time.Time{}, errors.New("диапазон должен быть 'dd dd' или 'dd.mm dd.mm'")
	}
}

func parseDayDay(a, b string, now time.Time, loc *time.Location) (time.Time, time.Time, error) {
	dayA, _ := strconv.Atoi(a)
	dayB, _ := strconv.Atoi(b)

	start := time.Date(now.Year(), now.Month(), dayA, 8, 0, 0, 0, loc)

	var end time.Time
	if dayB >= dayA {
		// the same month
		end = time.Date(now.Year(), now.Month(), dayB, 22, 0, 0, 0, loc)
	} else {
		// the next month
		year, month := now.Year(), now.Month()+1
		if month > 12 {
			month = 1
			year++
		}
		end = time.Date(year, month, dayB, 22, 0, 0, 0, loc)
	}
	return start, end, nil
}

func parseDDMM(token string, year int, loc *time.Location) (time.Time, error) {
	m := reDayMonth.FindStringSubmatch(token)
	if len(m) != 3 {
		return time.Time{}, errors.New("ожидал dd.mm")
	}
	day, _ := strconv.Atoi(m[1])
	mon, _ := strconv.Atoi(m[2])

	return time.Date(year, time.Month(mon), day, 8, 0, 0, 0, loc), nil
}

func parseDDMM_DDMM(a, b string, now time.Time, loc *time.Location) (time.Time, time.Time, error) {
	// Try current year
	start, err := parseDDMM(a, now.Year(), loc)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	end, err := parseDDMM(b, now.Year(), loc)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	// If both dates before Now
	today := truncateToDay(now.In(loc))
	if end.Before(today) && start.Before(today) {
		start, err = parseDDMM(a, now.Year()+1, loc)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		end, err = parseDDMM(b, now.Year()+1, loc)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	// If the end before start
	if end.Before(start) {
		// TODO: обмозговать
		end = start.Add(24 * time.Hour)
	}

	end = time.Date(end.Year(), end.Month(), end.Day(), 22, 0, 0, 0, loc)
	return start, end, nil
}

func truncateToDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}
