package memory

import (
	"context"
	"sync"
	"time"

	domain "github.com/right1121/railway-control-center-simulator/internal/domain/session"
)

// InMemorySessionRepository は単一 TrainingSession をメモリで保持するRepository実装。
type InMemorySessionRepository struct {
	mu      sync.Mutex
	session *domain.TrainingSession
}

// InMemorySessionRepository はリポジトリを生成する。
func NewInMemorySessionRepository() domain.Repository {
	return &InMemorySessionRepository{}
}

// Get はセッションを取得する。存在しなければ生成する。
func (r *InMemorySessionRepository) Get(ctx context.Context) (*domain.TrainingSession, error) {
	_ = ctx

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.session == nil {
		// 単一セッションID（あとでルーム化するならここを差し替える）
		id, err := domain.NewSessionID("default")
		if err != nil {
			return nil, err
		}
		r.session = domain.NewTrainingSession(id, time.Now())
	}
	return r.session, nil
}

func (r *InMemorySessionRepository) Create(ctx context.Context, s *domain.TrainingSession) error {
	_ = ctx

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.session != nil {
		return domain.ErrSessionAlreadyExists // ←後述：エラーを追加する or CONFLICTにする
	}
	r.session = s
	return nil
}

func (r *InMemorySessionRepository) Save(ctx context.Context, s *domain.TrainingSession) error {
	_ = ctx

	r.mu.Lock()
	defer r.mu.Unlock()

	r.session = s
	return nil
}
