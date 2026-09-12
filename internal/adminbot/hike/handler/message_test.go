package handler

import (
	"strings"
	"testing"

	"github.com/boris-guzeev/aktiv-hike-bot/internal/adminbot/hike/service"
)

func TestBuildHikeDetailsMessageIncludesType(t *testing.T) {
	hike := service.Hike{
		ID:      1,
		TitleRu: "Хайк",
		Type: &service.HikeType{
			ID:     8,
			Name:   "Волонтёрство / помощь на мероприятии AktivHike",
			Points: 1,
		},
	}

	message := buildHikeDetailsMessage(hike)
	want := "<b>Тип:</b> Волонтёрство / помощь на мероприятии AktivHike (1 балл)"
	if !strings.Contains(message, want) {
		t.Fatalf("message does not contain %q:\n%s", want, message)
	}
}

func TestPointsWord(t *testing.T) {
	tests := map[int32]string{
		1:  "балл",
		2:  "балла",
		5:  "баллов",
		11: "баллов",
		21: "балл",
	}

	for points, want := range tests {
		if got := pointsWord(points); got != want {
			t.Errorf("pointsWord(%d) = %q, want %q", points, got, want)
		}
	}
}
