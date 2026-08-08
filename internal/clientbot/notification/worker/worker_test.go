package worker

import "testing"

func TestCompletedText(t *testing.T) {
	distance := 12.5
	elevation := int32(850)

	got := completedText(bookingStatusChangedPayload{
		HikeTitle:     "Казбеги",
		DistanceKm:    &distance,
		ElevationGain: &elevation,
	})
	want := "Поздравляем! Вы завершили хайк \"Казбеги\" 🎉\n" +
		"🥾 Пройдено: 12.5 км\n" +
		"⛰ Набор высоты: 850 м\n\n" +
		"Отличное достижение! Так держать 🙌"

	if got != want {
		t.Fatalf("unexpected message:\n%s", got)
	}
}
