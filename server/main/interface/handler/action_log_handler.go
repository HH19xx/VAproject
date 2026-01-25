package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"server/main/usecases"
)

type ActionLogHandler struct {
	Usecase *usecases.ActionLogUsecase
}

// GET /api/v1/action_logs?target_id=&page=&limit=
func (h *ActionLogHandler) ListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET のみ許可されています", nil)
			return
		}
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit < 1 || limit > 100 {
			limit = 20
		}
		targetID, _ := strconv.Atoi(r.URL.Query().Get("target_id"))

		logs, total, err := h.Usecase.GetActionLogs(r.Context(), targetID, page, limit)
		if err != nil {
			JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "行動ログの取得に失敗しました", err.Error())
			return
		}
		JSONSuccess(w, http.StatusOK, map[string]interface{}{
			"logs": logs,
			"pagination": map[string]interface{}{
				"page":  page,
				"limit": limit,
				"total": total,
			},
		}, "行動ログを取得しました")
	}
}

// POST /api/v1/action_logs
func (h *ActionLogHandler) CreateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "POST のみ許可されています", nil)
			return
		}
		var req struct {
			TargetID   *int   `json:"target_id"`
			ActionType int    `json:"action_type"`
			Timestamp  string `json:"timestamp"` // ISO8601 (RFC3339)
			Notes      string `json:"notes"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			JSONError(w, http.StatusBadRequest, "INVALID_JSON", "JSON の解析に失敗しました", err.Error())
			return
		}
		if req.TargetID != nil && *req.TargetID <= 0 {
			JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", usecases.ErrActionLogTargetInvalid.Error(), nil)
			return
		}
		ts, err := time.Parse(time.RFC3339, req.Timestamp)
		if err != nil {
			JSONError(w, http.StatusBadRequest, "INVALID_TIMESTAMP", "timestamp は RFC3339 形式で指定してください", nil)
			return
		}
		username, _ := r.Context().Value("username").(string)
		if username == "" {
			username = "system"
		}
		log, err := h.Usecase.CreateActionLog(r.Context(), req.TargetID, req.ActionType, ts, req.Notes, username)
		if err != nil {
			switch err {
			case usecases.ErrActionLogTargetInvalid, usecases.ErrActionLogTypeRequired, usecases.ErrActionLogTimeRequired:
				JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
				return
			default:
				JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "行動ログの作成に失敗しました", err.Error())
				return
			}
		}
		JSONSuccess(w, http.StatusCreated, log, "行動ログを作成しました")
	}
}

// GET/PUT/DELETE /api/v1/action_logs/{id}
func (h *ActionLogHandler) DetailHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/action_logs/")
		idStr := strings.Split(path, "/")[0]
		id, err := strconv.Atoi(idStr)
		if err != nil || id < 1 {
			JSONError(w, http.StatusBadRequest, "INVALID_ID", "ID が不正です", nil)
			return
		}
		switch r.Method {
		case http.MethodGet:
			h.getByID(w, r, id)
		case http.MethodPut:
			h.update(w, r, id)
		case http.MethodDelete:
			h.delete(w, r, id)
		default:
			JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET/PUT/DELETE のみ許可されています", nil)
		}
	}
}

func (h *ActionLogHandler) getByID(w http.ResponseWriter, r *http.Request, id int) {
	log, err := h.Usecase.GetActionLogByID(r.Context(), id)
	if err != nil {
		JSONError(w, http.StatusNotFound, "NOT_FOUND", "行動ログが見つかりませんでした", nil)
		return
	}
	JSONSuccess(w, http.StatusOK, log, "行動ログを取得しました")
}

func (h *ActionLogHandler) update(w http.ResponseWriter, r *http.Request, id int) {
	var req struct {
		TargetID   *int   `json:"target_id"`
		ActionType int    `json:"action_type"`
		Timestamp  string `json:"timestamp"`
		Notes      string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_JSON", "JSON の解析に失敗しました", err.Error())
		return
	}
	if req.TargetID != nil && *req.TargetID <= 0 {
		JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", usecases.ErrActionLogTargetInvalid.Error(), nil)
		return
	}
	ts, err := time.Parse(time.RFC3339, req.Timestamp)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_TIMESTAMP", "timestamp は RFC3339 形式で指定してください", nil)
		return
	}
	username, _ := r.Context().Value("username").(string)
	if username == "" {
		username = "system"
	}
	log, err := h.Usecase.UpdateActionLog(r.Context(), id, req.TargetID, req.ActionType, ts, req.Notes, username)
	if err != nil {
		switch err {
		case usecases.ErrActionLogTargetInvalid, usecases.ErrActionLogTypeRequired, usecases.ErrActionLogTimeRequired:
			JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			return
		default:
			JSONError(w, http.StatusNotFound, "NOT_FOUND", "行動ログが見つかりませんでした", nil)
			return
		}
	}
	JSONSuccess(w, http.StatusOK, log, "行動ログを更新しました")
}

func (h *ActionLogHandler) delete(w http.ResponseWriter, r *http.Request, id int) {
	if err := h.Usecase.DeleteActionLog(r.Context(), id); err != nil {
		JSONError(w, http.StatusNotFound, "NOT_FOUND", "行動ログが見つかりませんでした", nil)
		return
	}
	JSONSuccess(w, http.StatusOK, nil, "行動ログを削除しました")
}
