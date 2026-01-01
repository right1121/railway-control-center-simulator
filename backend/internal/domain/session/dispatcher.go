package session

import "time"

// Dispatcher は訓練セッションに参加する管制員（Entity）
// IDで同一性を判断する。
type Dispatcher struct {
	id       DispatcherID
	name     DispatcherName
	joinedAt time.Time
}

// NewDispatcher は管制員を生成する（生成時に不変条件をチェック）
func NewDispatcher(id DispatcherID, name DispatcherName) Dispatcher {
	return Dispatcher{
		id:   id,
		name: name,
	}
}

func (d *Dispatcher) joined(t time.Time) {
	d.joinedAt = t
}

func (d Dispatcher) ID() DispatcherID     { return d.id }
func (d Dispatcher) Name() DispatcherName { return d.name }
func (d Dispatcher) JoinedAt() time.Time  { return d.joinedAt }
