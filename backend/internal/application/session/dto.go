package session

import (
	"time"

	domain "github.com/right1121/railway-control-center-simulator/internal/domain/session"
)

type SessionSnapshotDTO struct {
	SessionID   string          `json:"sessionId"`
	Dispatchers []DispatcherDTO `json:"dispatchers"`
}

type DispatcherDTO struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	JoinedAt time.Time `json:"joinedAt"`
}

// toSnapshotDTO はドメインからDTOへ変換する（application層の責務）
func toSnapshotDTO(s *domain.TrainingSession) SessionSnapshotDTO {
	dispatchers := s.Dispatchers()
	out := make([]DispatcherDTO, 0, len(dispatchers))
	for _, d := range dispatchers {
		out = append(out, DispatcherDTO{
			ID:       d.ID().String(),
			Name:     d.Name().String(),
			JoinedAt: d.JoinedAt(),
		})
	}

	return SessionSnapshotDTO{
		SessionID:   s.ID().String(),
		Dispatchers: out,
	}
}
