package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

var distancePattern = regexp.MustCompile(`^\d+(?:[.,]\d{1,2})?$`)

var requiredColumns = []string{
	"tg_username",
	"initial_distance_km",
	"initial_elevation_gain_m",
	"initial_points",
	"initial_events",
	"initial_achievements",
}

type userStat struct {
	Row            int
	Username       string
	DistanceKm     string
	ElevationGainM int32
	Points         int32
	Events         int32
	Achievements   []string
}

type importIssue struct {
	Row     int
	Message string
}
type parseResult struct {
	Stats    []userStat
	Issues   []importIssue
	RowsRead int
}

func parseUserStats(r io.Reader) (parseResult, error) {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		return parseResult{}, fmt.Errorf("read CSV: %w", err)
	}
	if len(records) == 0 {
		return parseResult{}, fmt.Errorf("CSV is empty")
	}
	columns, err := parseHeader(records[0])
	if err != nil {
		return parseResult{}, err
	}

	result := parseResult{Stats: make([]userStat, 0, len(records)-1)}
	seenUsernames := make(map[string]int)
	for i, record := range records[1:] {
		row := i + 2
		if isEmptyRow(record) {
			continue
		}
		result.RowsRead++
		username := normalizeUsername(field(record, columns["tg_username"]))
		if username == "" {
			result.Issues = append(result.Issues, importIssue{Row: row, Message: "tg_username is empty"})
			continue
		}
		if previousRow, ok := seenUsernames[username]; ok {
			result.Issues = append(result.Issues, importIssue{Row: row, Message: fmt.Sprintf("duplicate tg_username %q (first seen at row %d)", username, previousRow)})
			continue
		}
		distance, err := parseDistance(field(record, columns["initial_distance_km"]))
		if err != nil {
			result.Issues = append(result.Issues, importIssue{Row: row, Message: fmt.Sprintf("initial_distance_km: %v", err)})
			continue
		}
		elevation, err := parseNonNegativeInt32(field(record, columns["initial_elevation_gain_m"]))
		if err != nil {
			result.Issues = append(result.Issues, importIssue{Row: row, Message: fmt.Sprintf("initial_elevation_gain_m: %v", err)})
			continue
		}
		points, err := parseNonNegativeInt32(field(record, columns["initial_points"]))
		if err != nil {
			result.Issues = append(result.Issues, importIssue{Row: row, Message: fmt.Sprintf("initial_points: %v", err)})
			continue
		}
		events, err := parseNonNegativeInt32(field(record, columns["initial_events"]))
		if err != nil {
			result.Issues = append(result.Issues, importIssue{Row: row, Message: fmt.Sprintf("initial_events: %v", err)})
			continue
		}

		seenUsernames[username] = row
		result.Stats = append(result.Stats, userStat{
			Row: row, Username: username, DistanceKm: distance, ElevationGainM: elevation,
			Points: points, Events: events, Achievements: parseAchievementList(field(record, columns["initial_achievements"])),
		})
	}
	return result, nil
}

func parseHeader(header []string) (map[string]int, error) {
	required := make(map[string]bool, len(requiredColumns))
	for _, name := range requiredColumns {
		required[name] = true
	}

	columns := make(map[string]int, len(requiredColumns))
	for i, value := range header {
		name := strings.TrimSpace(strings.TrimPrefix(value, "\ufeff"))
		if !required[name] {
			continue
		}
		if _, exists := columns[name]; exists {
			return nil, fmt.Errorf("header: duplicate column %q", name)
		}
		columns[name] = i
	}

	var missing []string
	for _, name := range requiredColumns {
		if _, exists := columns[name]; !exists {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("header: missing required columns: %s", strings.Join(missing, ", "))
	}
	return columns, nil
}

func field(record []string, index int) string {
	if index >= len(record) {
		return ""
	}
	return record[index]
}

func parseAchievementList(value string) []string {
	var result []string
	for _, name := range strings.Split(value, ",") {
		name = strings.Join(strings.Fields(name), " ")
		if name != "" {
			result = append(result, name)
		}
	}
	return result
}

func isEmptyRow(record []string) bool {
	for _, value := range record {
		if strings.TrimSpace(value) != "" {
			return false
		}
	}
	return true
}

func normalizeUsername(value string) string {
	return strings.ToLower(strings.TrimPrefix(strings.TrimSpace(value), "@"))
}

func parseDistance(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "0", nil
	}
	if !distancePattern.MatchString(value) {
		return "", fmt.Errorf("must be a non-negative number with at most 2 decimal places")
	}
	value = strings.Replace(value, ",", ".", 1)
	distance, err := strconv.ParseFloat(value, 64)
	if err != nil || distance > 99999999.99 {
		return "", fmt.Errorf("value is outside NUMERIC(10,2) range")
	}
	return value, nil
}

func parseNonNegativeInt32(value string) (int32, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil || parsed < 0 {
		return 0, fmt.Errorf("must be a non-negative 32-bit integer")
	}
	return int32(parsed), nil
}
