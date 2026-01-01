package session

import "time"

// TrainingSession は訓練セッションの Aggregate Root
type TrainingSession struct {
	id           SessionID
	lastActiveAt time.Time
	dispatchers  map[string]Dispatcher // keyは DispatcherID.String()
	events       []DomainEvent
}

func NewTrainingSession(id SessionID, now time.Time) *TrainingSession {
	return &TrainingSession{
		id:           id,
		lastActiveAt: now,
		dispatchers:  make(map[string]Dispatcher),
		events:       make([]DomainEvent, 0),
	}
}

func (s *TrainingSession) ID() SessionID {
	return s.id
}

func (s *TrainingSession) Dispatchers() []Dispatcher {
	out := make([]Dispatcher, 0, len(s.dispatchers))
	for _, d := range s.dispatchers {
		out = append(out, d)
	}
	return out
}

func (s *TrainingSession) JoinDispatcher(
	dispatcher Dispatcher,
	now time.Time,
) error {
	dispatcher.joined(now)

	key := dispatcher.ID().String()

	if _, ok := s.dispatchers[key]; ok {
		return ErrDispatcherAlreadyExists
	}

	s.dispatchers[key] = dispatcher
	s.lastActiveAt = now

	s.events = append(
		s.events,
		NewDispatcherJoined(now, dispatcher),
	)
	return nil
}

func (s *TrainingSession) LeaveDispatcher(
	id DispatcherID,
	now time.Time,
) error {
	key := id.String()

	if _, ok := s.dispatchers[key]; !ok {
		return ErrDispatcherNotFound
	}

	delete(s.dispatchers, key)
	s.lastActiveAt = now

	s.events = append(
		s.events,
		NewDispatcherLeft(now, id),
	)
	return nil
}

func (s *TrainingSession) PullEvents() []DomainEvent {
	ev := s.events
	s.events = make([]DomainEvent, 0)
	return ev
}
