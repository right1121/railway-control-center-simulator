package session

import (
	"context"
	"errors"
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
	repo domain.Repository
	mu   sync.Mutex
}

// NewUseCase は UseCase 実装を生成する
func NewUseCase(repo domain.Repository) UseCase {
	return &service{
		repo: repo,
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
	session, err := s.ensureSession(ctx, now)
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

// ensureSession は「無ければ作る」をUseCaseの明示的な責務として実装する
func (s *service) ensureSession(ctx context.Context, now time.Time) (*domain.TrainingSession, error) {
	session, err := s.repo.Get(ctx)
	if err == nil {
		return session, nil
	}

	if !errors.Is(err, domain.ErrSessionAlreadyExists) {
		return nil, err
	}

	// 無ければ明示的に作る
	id, err := domain.NewSessionID("default")
	if err != nil {
		return nil, fmt.Errorf("セッションIDの生成に失敗。%w", err)
	}
	newSession := domain.NewTrainingSession(id, now)

	// Create（競合したら再Getで救済）
	if err := s.repo.Create(ctx, newSession); err != nil {
		// 同時リクエストで二重作成になった場合を救済
		// （service.mu があるなら基本起きないが、repoが別プロセス等になっても安全）
		if errors.Is(err, domain.ErrSessionAlreadyExists) {
			again, gerr := s.repo.Get(ctx)
			if gerr == nil {
				return again, nil
			}
			return nil, fmt.Errorf("セッション作成競合後の取得に失敗: %w", err)
		}
		return nil, fmt.Errorf("セッションの作成に失敗: %w", err)
	}

	return newSession, nil
}
