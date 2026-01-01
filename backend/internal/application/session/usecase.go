package session

import (
	"context"
	"fmt"
	"sync"
	"time"

	domain "github.com/right1121/railway-control-center-simulator/internal/domain/session"
)

type UseCase interface {
	JoinDispatcher(ctx context.Context, input JoinDispatcherInput) (JoinDispatcherOutput, error)
	LeaveDispatcher(ctx context.Context, input LeaveDispatcherInput) error
	GetSnapshot(ctx context.Context) (SessionSnapshotDTO, error)
}

// JoinDispatcherInput は参加ユースケースの入力
type JoinDispatcherInput struct {
	DispatcherID domain.DispatcherID
	Name         domain.DispatcherName
	Now          time.Time
}

// JoinDispatcherOutput は参加ユースケースの出力
type JoinDispatcherOutput struct {
	DispatcherID string             `json:"dispatcherId"`
	Snapshot     SessionSnapshotDTO `json:"snapshot"`
}

// LeaveDispatcherInput は退出ユースケースの入力
type LeaveDispatcherInput struct {
	DispatcherID string
	Now          time.Time
}

// service は UseCase の実装
// - repo 経由でドメインを取得・保存
// - mutex で整合性（複数操作の直列化）を保証
type service struct {
	repo domain.TrainingSessionRepository
	mu   sync.Mutex

	// ID生成関数は差し替え可能にしてテストしやすくする
	newDispatcherID func() string
}

// NewUseCase は UseCase 実装を生成する
func NewUseCase(repo domain.TrainingSessionRepository, newDispatcherID func() string) UseCase {
	if newDispatcherID == nil {
		newDispatcherID = func() string { return "TODO" } // 実運用では必ず差し替える
	}
	return &service{
		repo:            repo,
		newDispatcherID: newDispatcherID,
	}
}

// JoinDispatcher は管理者をセッションに参加させます
func (s *service) JoinDispatcher(ctx context.Context, input JoinDispatcherInput) (JoinDispatcherOutput, error) {
	// すべての更新系ユースケースは直列化する（単一セッションの整合性確保）
	s.mu.Lock()
	defer s.mu.Unlock()

	now := input.Now
	if now.IsZero() {
		now = time.Now()
	}

	dispatcher := domain.NewDispatcher(
		input.DispatcherID,
		input.Name,
	)

	// ドメイン集約取得
	session, err := s.repo.Get(ctx)
	if err != nil {
		return JoinDispatcherOutput{}, err
	}

	// ドメイン操作
	if err := session.JoinDispatcher(dispatcher, now); err != nil {
		return JoinDispatcherOutput{}, fmt.Errorf("ディスパッチャーの参加に失敗: %w", err)
	}

	// 保存
	if err := s.repo.Save(ctx, session); err != nil {
		return JoinDispatcherOutput{}, fmt.Errorf("セッションの保存に失敗: %w", err)
	}

	// スナップショット生成（DTO変換）
	snapshot := toSnapshotDTO(session)

	return JoinDispatcherOutput{
		DispatcherID: dispatcher.ID().String(),
		Snapshot:     snapshot,
	}, nil
}

func (s *service) LeaveDispatcher(ctx context.Context, input LeaveDispatcherInput) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := input.Now
	if now.IsZero() {
		now = time.Now()
	}

	idVO, err := domain.NewDispatcherID(input.DispatcherID)
	if err != nil {
		return err
	}

	session, err := s.repo.Get(ctx)
	if err != nil {
		return err
	}

	if err := session.LeaveDispatcher(idVO, now); err != nil {
		switch err {
		case domain.ErrDispatcherNotFound:
			return err
		default:
			return err
		}
	}

	if err := s.repo.Save(ctx, session); err != nil {
		return err
	}

	return nil
}

func (s *service) GetSnapshot(ctx context.Context) (SessionSnapshotDTO, error) {
	// 読み取りでも、repoがメモリ参照ならロック不要にできるが、
	// まずは安全側（更新と競合させない）で mu を使うのが無難。
	s.mu.Lock()
	defer s.mu.Unlock()

	session, err := s.repo.Get(ctx)
	if err != nil {
		return SessionSnapshotDTO{}, err
	}
	return toSnapshotDTO(session), nil
}
