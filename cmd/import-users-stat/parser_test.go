package main

import (
	"strings"
	"testing"
)

const testHeader = "tg_username,initial_distance_km,initial_elevation_gain_m,initial_points,initial_events,initial_achievements\n"

func TestParseUserStats(t *testing.T) {
	csv := testHeader + "@Alice,12.50,800,17,3,\"4000+, III SEASONS, II SEASONS\"\n" + "bob,\"7,25\",400,8,2,TREKKER\n"
	result, err := parseUserStats(strings.NewReader(csv))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Stats) != 2 {
		t.Fatalf("got %d rows, want 2", len(result.Stats))
	}
	first := result.Stats[0]
	if first.Username != "alice" || first.DistanceKm != "12.50" || first.Points != 17 || first.Events != 3 {
		t.Fatalf("unexpected first row: %+v", first)
	}
	if len(first.Achievements) != 3 || first.Achievements[1] != "III SEASONS" {
		t.Fatalf("unexpected achievements: %#v", first.Achievements)
	}
	if result.Stats[1].DistanceKm != "7.25" {
		t.Fatalf("unexpected second row: %+v", result.Stats[1])
	}
}

func TestParseUserStatsFindsColumnsByNameAndIgnoresOthers(t *testing.T) {
	csv := "name,initial_events,notes,initial_achievements,tg_username,initial_points,initial_elevation_gain_m,unused,initial_distance_km\n" +
		"Alice,3,legacy,\"TREKKER, III SEASONS\",@Alice,17,800,anything,12.50\n"
	result, err := parseUserStats(strings.NewReader(csv))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Stats) != 1 {
		t.Fatalf("got %d rows, want 1", len(result.Stats))
	}
	got := result.Stats[0]
	if got.Username != "alice" || got.DistanceKm != "12.50" || got.ElevationGainM != 800 || got.Points != 17 || got.Events != 3 {
		t.Fatalf("unexpected row: %+v", got)
	}
	if len(got.Achievements) != 2 || got.Achievements[1] != "III SEASONS" {
		t.Fatalf("unexpected achievements: %#v", got.Achievements)
	}
}

func TestParseUserStatsReportsMissingRequiredColumn(t *testing.T) {
	csv := "tg_username,initial_distance_km,initial_elevation_gain_m,initial_points,initial_events\n"
	_, err := parseUserStats(strings.NewReader(csv))
	if err == nil || !strings.Contains(err.Error(), "initial_achievements") {
		t.Fatalf("got error %v, want missing initial_achievements", err)
	}
}

func TestParseUserStatsReportsDuplicateRequiredColumn(t *testing.T) {
	csv := "tg_username,initial_distance_km,initial_elevation_gain_m,initial_points,initial_events,initial_achievements,tg_username\n"
	_, err := parseUserStats(strings.NewReader(csv))
	if err == nil || !strings.Contains(err.Error(), "duplicate column") {
		t.Fatalf("got error %v, want duplicate column error", err)
	}
}

func TestParseUserStatsReportsDuplicateUsername(t *testing.T) {
	csv := testHeader + "Alice,12,800,1,3,\n@alice,15,900,2,4,\n"
	result, err := parseUserStats(strings.NewReader(csv))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Issues) != 1 || !strings.Contains(result.Issues[0].Message, "duplicate tg_username") {
		t.Fatalf("unexpected issues: %+v", result.Issues)
	}
}

func TestParseUserStatsReportsInvalidValue(t *testing.T) {
	csv := testHeader + "alice,12.345,800,1,3,\n"
	result, err := parseUserStats(strings.NewReader(csv))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Issues) != 1 || !strings.Contains(result.Issues[0].Message, "initial_distance_km") {
		t.Fatalf("unexpected issues: %+v", result.Issues)
	}
}

func TestParseUserStatsTreatsEmptyNumericCellsAsZero(t *testing.T) {
	csv := testHeader + "alice,,,,,\n"
	result, err := parseUserStats(strings.NewReader(csv))
	if err != nil {
		t.Fatal(err)
	}
	got := result.Stats[0]
	if got.DistanceKm != "0" || got.ElevationGainM != 0 || got.Points != 0 || got.Events != 0 {
		t.Fatalf("unexpected row: %+v", got)
	}
}

func TestParseUserStatsReportsMissingUsernameAndContinues(t *testing.T) {
	csv := testHeader + ",12,800,1,3,\nbob,10,500,2,2,\n"
	result, err := parseUserStats(strings.NewReader(csv))
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Stats) != 1 || result.Stats[0].Username != "bob" {
		t.Fatalf("unexpected stats: %+v", result.Stats)
	}
	if len(result.Issues) != 1 || result.Issues[0].Message != "tg_username is empty" {
		t.Fatalf("unexpected issues: %+v", result.Issues)
	}
}

func TestResolveAchievementsKeepsLargestSeasonAndUnknown(t *testing.T) {
	known := map[string]achievement{
		"4000+":   {ID: 1, Name: "4000+"},
		"SEASONS": {ID: 12, Name: "SEASONS", HasValue: true},
	}
	resolved, unknown := resolveAchievements([]string{"4000+", "II SEASONS", "5 SEASONS", "MYSTERY"}, known)
	if resolved[1] != 0 || resolved[12] != 5 {
		t.Fatalf("unexpected resolved achievements: %+v", resolved)
	}
	if len(unknown) != 1 || unknown[0] != "MYSTERY" {
		t.Fatalf("unexpected unknown achievements: %#v", unknown)
	}
}
