package session

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	sessionapp "github.com/right1121/railway-control-center-simulator/internal/application/session"
	domain "github.com/right1121/railway-control-center-simulator/internal/domain/session"
	"github.com/right1121/railway-control-center-simulator/internal/interfaces/http/utils"
	"github.com/right1121/railway-control-center-simulator/pkg/appctx"
)

type SessionHandler struct {
	usecase sessionapp.UseCase
}

func NewSessionHandler(uc sessionapp.UseCase) *SessionHandler {
	return &SessionHandler{usecase: uc}
}

func (h *SessionHandler) Get(w http.ResponseWriter, r *http.Request) {
	dto, err := h.usecase.GetSnapshot(r.Context())
	if err != nil {
		writeUseCaseError(w, err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, dto)
}

type joinReq struct {
	DispatcherID string `json:"dispatcherId"`
	Name         string `json:"name"`
}

func (h *SessionHandler) Join(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromRequest(r)
	logger := ctx.GetLogger()

	var req joinReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.WithError(fmt.Errorf("JSONの変換に失敗しました: %w", err))
		utils.WriteJSON(w, http.StatusBadRequest, utils.BadJSON())
		return
	}

	dispatcherID, err := domain.NewDispatcherID(req.DispatcherID)
	if err != nil {
		logger.WithError(fmt.Errorf("不正な管理者IDが指定されました: %w", err))
		utils.WriteJSON(w, http.StatusBadRequest, utils.ErrBody("INVALID_DISPATCHER_ID", "invalid dispatcher ID"))
		return
	}
	name, err := domain.NewDispatcherName(req.Name)
	if err != nil {
		logger.WithError(fmt.Errorf("不正な管理者名が指定されました: %w", err))
		utils.WriteJSON(w, http.StatusBadRequest, utils.ErrBody("INVALID_NAME", "invalid dispatcher name"))
		return
	}

	out, err := h.usecase.JoinDispatcher(r.Context(), sessionapp.JoinDispatcherInput{
		DispatcherID: dispatcherID,
		Name:         name,
		Now:          time.Now(),
	})
	if err != nil {
		logger.WithError(fmt.Errorf("管理者の参加に失敗しました: %w", err))
		writeUseCaseError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, out)
}

type leaveReq struct {
	DispatcherID string `json:"dispatcherId"`
}

func (h *SessionHandler) Leave(w http.ResponseWriter, r *http.Request) {
	var req leaveReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSON(w, http.StatusBadRequest, utils.BadJSON())
		return
	}

	if err := h.usecase.LeaveDispatcher(r.Context(), sessionapp.LeaveDispatcherInput{
		DispatcherID: req.DispatcherID,
		Now:          time.Now(),
	}); err != nil {
		writeUseCaseError(w, err)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func writeUseCaseError(w http.ResponseWriter, err error) {
	utils.WriteJSON(w, http.StatusInternalServerError, utils.ErrBody("INTERNAL", "internal error"))
}
