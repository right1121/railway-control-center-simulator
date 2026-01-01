package session

import (
	"context"
)

// Repository はドメイン集約の永続化境界
type Repository interface {
	// Get は単一セッションを取得する
	Get(ctx context.Context) (*TrainingSession, error)

	// Create は新規セッションを作成する
	Create(ctx context.Context, session *TrainingSession) error

	// Save は変更されたセッションを保存する
	Save(ctx context.Context, s *TrainingSession) error
}
