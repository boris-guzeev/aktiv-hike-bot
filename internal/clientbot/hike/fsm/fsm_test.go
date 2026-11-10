package fsm

import "testing"

func TestFinishActualHikeSelectionKeepsDisplayedHikes(t *testing.T) {
	f := New()
	const userID int64 = 42

	f.SetActualHikes(userID, []int32{101, 202})
	f.FinishActualHikeSelection(userID)

	if state := f.GetState(userID); state != StateIdle {
		t.Fatalf("state = %q, want %q", state, StateIdle)
	}
	if !f.HasActualHikes(userID) {
		t.Fatal("displayed hikes were removed after selection")
	}
	if hikeID, ok := f.GetHikeID(userID, 2); !ok || hikeID != 202 {
		t.Fatalf("GetHikeID(2) = (%d, %v), want (202, true)", hikeID, ok)
	}
}

func TestResetRemovesDisplayedHikes(t *testing.T) {
	f := New()
	const userID int64 = 42

	f.SetActualHikes(userID, []int32{101})
	f.Reset(userID)

	if f.HasActualHikes(userID) {
		t.Fatal("Reset kept displayed hikes")
	}
}
