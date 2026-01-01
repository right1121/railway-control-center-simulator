package session

import (
	"context"
)

// TrainingSessionRepository はドメイン集約の永続化境界
type TrainingSessionRepository interface {
	// Get は単一セッションを取得する
	Get(ctx context.Context) (*TrainingSession, error)

	// Save は変更されたセッションを保存する
	Save(ctx context.Context, s *TrainingSession) error
}
