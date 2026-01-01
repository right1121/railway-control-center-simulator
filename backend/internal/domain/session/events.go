package session

import "time"

type DomainEvent interface {
	EventType() string
	OccurredAt() time.Time
}

type DispatcherJoined struct {
	at   time.Time
	id   DispatcherID
	name DispatcherName
}

func NewDispatcherJoined(at time.Time, dispatcher Dispatcher) DispatcherJoined {
	return DispatcherJoined{
		at:   at,
		id:   dispatcher.ID(),
		name: dispatcher.Name(),
	}
}

func (e DispatcherJoined) EventType() string {
	return "DISPATCHER_JOINED"
}

func (e DispatcherJoined) OccurredAt() time.Time {
	return e.at
}

type DispatcherLeft struct {
	at time.Time
	id DispatcherID
}

func NewDispatcherLeft(at time.Time, id DispatcherID) DispatcherLeft {
	return DispatcherLeft{
		at: at,
		id: id,
	}
}

func (e DispatcherLeft) EventType() string {
	return "DISPATCHER_LEFT"
}

func (e DispatcherLeft) OccurredAt() time.Time {
	return e.at
}
