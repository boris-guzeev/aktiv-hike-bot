package hike

import "testing"

func TestListHikesPaginationKeyboard(t *testing.T) {
	tests := []struct {
		name     string
		page     int32
		hasNext  bool
		wantData []string
	}{
		{name: "single page", page: 1, wantData: nil},
		{name: "first page", page: 1, hasNext: true, wantData: []string{"hike:list:page:2"}},
		{name: "middle page", page: 2, hasNext: true, wantData: []string{"hike:list:page:1", "hike:list:page:3"}},
		{name: "last page", page: 3, wantData: []string{"hike:list:page:2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyboard := ListHikesPaginationKeyboard(tt.page, tt.hasNext)
			if len(tt.wantData) == 0 {
				if len(keyboard.InlineKeyboard) != 0 {
					t.Fatalf("got %d rows, want none", len(keyboard.InlineKeyboard))
				}
				return
			}

			if len(keyboard.InlineKeyboard) != 1 {
				t.Fatalf("got %d rows, want 1", len(keyboard.InlineKeyboard))
			}
			row := keyboard.InlineKeyboard[0]
			if len(row) != len(tt.wantData) {
				t.Fatalf("got %d buttons, want %d", len(row), len(tt.wantData))
			}
			for i, want := range tt.wantData {
				if row[i].CallbackData == nil || *row[i].CallbackData != want {
					t.Errorf("button %d callback = %v, want %q", i, row[i].CallbackData, want)
				}
			}
		})
	}
}
