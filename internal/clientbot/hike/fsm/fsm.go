package fsm

import "sync"

type State string

var (
	StateIdle State = "idle"

	StateWaitingActualHikeNumber State = "waiting_actual_hike_number"
)

type FSM struct {
	mu          sync.Mutex
	userState   map[int64]State
	userHikeIDs map[int64][]int32
}

func New() *FSM {
	return &FSM{
		userState:   make(map[int64]State),
		userHikeIDs: make(map[int64][]int32),
	}
}

func (f *FSM) SetActualHikes(userID int64, ids []int32) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.userState[userID] = StateWaitingActualHikeNumber
	f.userHikeIDs[userID] = ids
}

// func (f *FSM) SetState(userID int64, state State) {
// 	f.mu.Lock()
// 	defer f.mu.Unlock()

// 	f.userState[userID] = state
// }

func (f *FSM) GetHikeID(userID int64, number int32) (int32, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	list, ok := f.userHikeIDs[userID]
	if !ok {
		return 0, false
	}

	if number < 1 || number > int32(len(list)) {
		return 0, false
	}

	return list[number-1], true
}

func (f *FSM) HasActualHikes(userID int64) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	return len(f.userHikeIDs[userID]) > 0
}

// FinishActualHikeSelection ends the number prompt but keeps the displayed
// list, so another number can select a different hike from the same list.
func (f *FSM) FinishActualHikeSelection(userID int64) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.userState[userID] = StateIdle
}

func (f *FSM) GetState(userID int64) State {
	f.mu.Lock()
	defer f.mu.Unlock()

	state, ok := f.userState[userID]
	if !ok {
		return StateIdle
	}
	return state
}

func (f *FSM) Reset(userID int64) {
	f.mu.Lock()
	defer f.mu.Unlock()

	delete(f.userState, userID)
	delete(f.userHikeIDs, userID)
}
