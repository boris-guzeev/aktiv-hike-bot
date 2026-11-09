package main

import (
	"strings"
	"testing"
)

func TestParseUserStats(t *testing.T) {
	csv := `name,notes,tg_username,initial_distance_km,initial_elevation_gain_m,initial_hikes_count
Alice,,@Alice,12.50,800,3
Bob,,bob,"7,25",400,2
`

	result, err := parseUserStats(strings.NewReader(csv))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Stats) != 2 {
		t.Fatalf("got %d rows, want 2", len(result.Stats))
	}
	if result.Stats[0].Username != "alice" || result.Stats[0].DistanceKm != "12.50" {
		t.Fatalf("unexpected first row: %+v", result.Stats[0])
	}
	if result.Stats[1].Username != "bob" || result.Stats[1].DistanceKm != "7.25" {
		t.Fatalf("unexpected second row: %+v", result.Stats[1])
	}
}

func TestParseUserStatsReportsDuplicateUsername(t *testing.T) {
	csv := `name,notes,tg_username,initial_distance_km,initial_elevation_gain_m,initial_hikes_count
Alice,,Alice,12,800,3
Alice again,,@alice,15,900,4
`

	result, err := parseUserStats(strings.NewReader(csv))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Issues) != 1 || !strings.Contains(result.Issues[0].Message, "duplicate tg_username") {
		t.Fatalf("got issues %+v, want duplicate username issue", result.Issues)
	}
}

func TestParseUserStatsReportsInvalidValue(t *testing.T) {
	csv := `name,notes,tg_username,initial_distance_km,initial_elevation_gain_m,initial_hikes_count
Alice,,alice,12.345,800,3
`

	result, err := parseUserStats(strings.NewReader(csv))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Issues) != 1 || !strings.Contains(result.Issues[0].Message, "initial_distance_km") {
		t.Fatalf("got issues %+v, want distance issue", result.Issues)
	}
}

func TestParseUserStatsAcceptsIntegerAndZeroDistance(t *testing.T) {
	csv := `name,notes,tg_username,initial_distance_km,initial_elevation_gain_m,initial_hikes_count
Alice,,alice,0,0,0
Bob,,bob,12,800,3
`

	result, err := parseUserStats(strings.NewReader(csv))
	if err != nil {
		t.Fatal(err)
	}
	if result.Stats[0].DistanceKm != "0" {
		t.Fatalf("got zero distance %q, want 0", result.Stats[0].DistanceKm)
	}
	if result.Stats[1].DistanceKm != "12" {
		t.Fatalf("got integer distance %q, want 12", result.Stats[1].DistanceKm)
	}
}

func TestParseUserStatsTreatsEmptyNumericCellsAsZero(t *testing.T) {
	csv := `name,notes,tg_username,initial_distance_km,initial_elevation_gain_m,initial_hikes_count
Alice,,alice,,,
`

	result, err := parseUserStats(strings.NewReader(csv))
	if err != nil {
		t.Fatal(err)
	}
	if result.Stats[0].DistanceKm != "0" || result.Stats[0].ElevationGainM != 0 || result.Stats[0].HikesCount != 0 {
		t.Fatalf("unexpected parsed row: %+v", result.Stats[0])
	}
}

func TestParseUserStatsReportsMissingUsernameAndContinues(t *testing.T) {
	csv := `name,notes,tg_username,initial_distance_km,initial_elevation_gain_m,initial_hikes_count
Alice,,,12,800,3
Bob,,bob,10,500,2
`

	result, err := parseUserStats(strings.NewReader(csv))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Stats) != 1 || result.Stats[0].Username != "bob" {
		t.Fatalf("unexpected valid stats: %+v", result.Stats)
	}
	if len(result.Issues) != 1 || result.Issues[0].Message != "tg_username is empty" {
		t.Fatalf("unexpected issues: %+v", result.Issues)
	}
}
