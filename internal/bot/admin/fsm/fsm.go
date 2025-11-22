package fsm

import "sync"

type State string

const (
	StateIdle State = "idle"
	StateCreateTitleRU State = "create_title_ru"
	StateCreateTitleEN State = "create_title_en"
	StateCreateDescRU State = "create_desc_ru"
	StateCreateDescEN State = "create_desc_en"
	StateCreateDates State = "create_dates"
	StateConfirm State = "confirm"
)

type session struct {
	State State
	Data map[string]string
}

type FSM struct {
	mu sync.Mutex
	userSessions map[int64]*session
}

func NewFSM() *FSM {
	return &FSM{
		userSessions: make(map[int64]*session),
	}
}

func (f *FSM) State(userID int64) State {
	f.mu.Lock()
	defer f.mu.Unlock()

	if s, ok := f.userSessions[userID]; ok {
		return s.State
	}
	return StateIdle
}

func (f *FSM) Set(userID int64, st State) {
	f.mu.Lock()
	defer f.mu.Unlock()

	s, ok := f.userSessions[userID]
	if !ok {
		s = &session{
			Data: map[string]string{},	
		}
		f.userSessions[userID] = s
	}
	s.State = st
}

func (f *FSM) Put(userID int64, k, v string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	s, ok := f.userSessions[userID]
	if !ok {
		s = &session{
			Data: map[string]string{},
		}
		f.userSessions[userID] = s
	}
	s.Data[k] = v
}

func (f *FSM) Data(userID int64) map[string]string {
	f.mu.Lock()
	defer f.mu.Unlock()

	s, ok := f.userSessions[userID]
	if !ok {
		return map[string]string{}
	}
	cp := make(map[string]string, len(s.Data))
	for k, v := range s.Data {
		cp[k] = v
	}
	return cp
}

func (f *FSM) Reset(userID int64) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.userSessions, userID)
}