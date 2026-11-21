package handler

import (
	"testing"

	userService "github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/user/service"
)

func TestToRoman(t *testing.T) {
	tests := map[int32]string{1: "I", 2: "II", 4: "IV", 9: "IX", 42: "XLII", 1999: "MCMXCIX", 0: ""}
	for value, want := range tests {
		if got := toRoman(value); got != want {
			t.Errorf("toRoman(%d) = %q, want %q", value, got, want)
		}
	}
}

func TestAchievementDisplayName(t *testing.T) {
	one, four := int32(1), int32(4)
	tests := []struct {
		achievement userService.Achievement
		want        string
	}{
		{userService.Achievement{Name: "SEASONS", Value: &one}, "I SEASON"},
		{userService.Achievement{Name: "SEASONS", Value: &four}, "IV SEASONS"},
		{userService.Achievement{Name: "SEASONS"}, "SEASONS"},
		{userService.Achievement{Name: "TREKKER", Value: &four}, "TREKKER"},
	}
	for _, tt := range tests {
		if got := achievementDisplayName(tt.achievement); got != tt.want {
			t.Errorf("achievementDisplayName(%+v) = %q, want %q", tt.achievement, got, tt.want)
		}
	}
}
