package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

const requiredColumnCount = 6

var distancePattern = regexp.MustCompile(`^\d+(?:[.,]\d{1,2})?$`)

type userStat struct {
	Row            int
	Username       string
	DistanceKm     string
	ElevationGainM int32
	HikesCount     int32
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
	if err := validateHeader(records[0]); err != nil {
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
		if len(record) < requiredColumnCount {
			result.Issues = append(result.Issues, importIssue{Row: row, Message: fmt.Sprintf("expected at least 6 columns, got %d", len(record))})
			continue
		}

		username := normalizeUsername(record[2])
		if username == "" {
			result.Issues = append(result.Issues, importIssue{Row: row, Message: "tg_username is empty"})
			continue
		}
		if previousRow, ok := seenUsernames[username]; ok {
			result.Issues = append(result.Issues, importIssue{Row: row, Message: fmt.Sprintf("duplicate tg_username %q (first seen at row %d)", username, previousRow)})
			continue
		}

		distance, err := parseDistance(record[3])
		if err != nil {
			result.Issues = append(result.Issues, importIssue{Row: row, Message: fmt.Sprintf("initial_distance_km: %v", err)})
			continue
		}
		elevation, err := parseNonNegativeInt32(record[4])
		if err != nil {
			result.Issues = append(result.Issues, importIssue{Row: row, Message: fmt.Sprintf("initial_elevation_gain_m: %v", err)})
			continue
		}
		hikes, err := parseNonNegativeInt32(record[5])
		if err != nil {
			result.Issues = append(result.Issues, importIssue{Row: row, Message: fmt.Sprintf("initial_hikes_count: %v", err)})
			continue
		}

		seenUsernames[username] = row
		result.Stats = append(result.Stats, userStat{
			Row:            row,
			Username:       username,
			DistanceKm:     distance,
			ElevationGainM: elevation,
			HikesCount:     hikes,
		})
	}

	return result, nil
}

func validateHeader(header []string) error {
	if len(header) < requiredColumnCount {
		return fmt.Errorf("header: expected at least 6 columns, got %d", len(header))
	}
	want := []string{
		"tg_username",
		"initial_distance_km",
		"initial_elevation_gain_m",
		"initial_hikes_count",
	}
	for i, expected := range want {
		actual := strings.TrimSpace(header[i+2])
		if actual != expected {
			return fmt.Errorf("header column %c: expected %q, got %q", 'C'+i, expected, actual)
		}
	}
	return nil
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
