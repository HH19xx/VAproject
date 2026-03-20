package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"server/main/usecases"
)

type TargetHandler struct {
	Usecase *usecases.TargetUsecase
}

// GET /api/v1/targets
func (h *TargetHandler) ListTargetsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET メソッドのみ許可されています", nil)
			return
		}

		userID, ok := currentUserID(r)
		if !ok {
			JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "認証が必要です", nil)
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

		targets, total, err := h.Usecase.GetTargets(r.Context(), userID, page, limit)
		if err != nil {
			JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "観察対象一覧の取得に失敗しました", err.Error())
			return
		}

		JSONSuccess(w, http.StatusOK, map[string]interface{}{
			"targets": targets,
			"pagination": map[string]interface{}{
				"page":  page,
				"limit": limit,
				"total": total,
			},
		}, "観察対象一覧を取得しました")
	}
}

// POST /api/v1/targets
func (h *TargetHandler) CreateTargetHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "POST メソッドのみ許可されています", nil)
			return
		}

		userID, ok := currentUserID(r)
		if !ok {
			JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "認証が必要です", nil)
			return
		}

		var req struct {
			Name          string  `json:"name"`
			Description   string  `json:"description"`
			MatchMode     string  `json:"match_mode"`
			QueryText     string  `json:"query_text"`
			TagIDs        []int   `json:"tag_ids"`
			AnyTagIDs     []int   `json:"any_tag_ids"`
			AnyTagGroups  [][]int `json:"any_tag_groups"`
			ExcludeTagIDs []int   `json:"exclude_tag_ids"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			JSONError(w, http.StatusBadRequest, "INVALID_JSON", "JSONの解析に失敗しました", err.Error())
			return
		}

		username, _ := r.Context().Value("username").(string)
		if username == "" {
			username = strconv.Itoa(userID)
		}

		target, err := h.Usecase.CreateTarget(
			r.Context(),
			userID,
			req.Name,
			req.Description,
			username,
			req.MatchMode,
			req.QueryText,
			req.TagIDs,
			req.AnyTagIDs,
			req.ExcludeTagIDs,
			req.AnyTagGroups,
		)
		if err != nil {
			switch {
			case errors.Is(err, usecases.ErrTargetNameRequired),
				errors.Is(err, usecases.ErrTargetNameTooLong),
				errors.Is(err, usecases.ErrTargetTagIDsRequired),
				errors.Is(err, usecases.ErrTargetTagIDInvalid),
				errors.Is(err, usecases.ErrTargetMatchModeInvalid):
				JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
				return
			case errors.Is(err, usecases.ErrTargetNameDuplicate):
				JSONError(w, http.StatusConflict, "CONFLICT", err.Error(), nil)
				return
			case errors.Is(err, usecases.ErrTargetTagNotFound):
				JSONError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
				return
			default:
				JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "観察対象の作成に失敗しました", err.Error())
				return
			}
		}

		JSONSuccess(w, http.StatusCreated, target, "観察対象を作成しました")
	}
}

// GET /api/v1/targets/{id}
// PUT /api/v1/targets/{id}
// DELETE /api/v1/targets/{id}
func (h *TargetHandler) TargetDetailHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/targets/")
		idStr := strings.Split(path, "/")[0]

		id, err := strconv.Atoi(idStr)
		if err != nil || id < 1 {
			JSONError(w, http.StatusBadRequest, "INVALID_ID", "IDが不正です", nil)
			return
		}

		switch r.Method {
		case http.MethodGet:
			h.getTargetByID(w, r, id)
		case http.MethodPut:
			h.updateTarget(w, r, id)
		case http.MethodDelete:
			h.deleteTarget(w, r, id)
		default:
			JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET/PUT/DELETE メソッドのみ許可されています", nil)
		}
	}
}

// GET /api/v1/targets/{id}/action_logs
func (h *TargetHandler) ListActionLogsByTargetHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET メソッドのみ許可されています", nil)
			return
		}
		userID, ok := currentUserID(r)
		if !ok {
			JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "認証が必要です", nil)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/api/v1/targets/")
		parts := strings.Split(path, "/")
		if len(parts) < 2 || parts[1] != "action_logs" {
			JSONError(w, http.StatusBadRequest, "INVALID_PATH", "パスが不正です", nil)
			return
		}
		targetID, err := strconv.Atoi(parts[0])
		if err != nil || targetID < 1 {
			JSONError(w, http.StatusBadRequest, "INVALID_ID", "IDが不正です", nil)
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

		logs, total, err := h.Usecase.GetActionLogsByTargetID(r.Context(), userID, targetID, page, limit)
		if err != nil {
			JSONError(w, http.StatusNotFound, "NOT_FOUND", "観察対象または行動記録が見つかりません", nil)
			return
		}

		JSONSuccess(w, http.StatusOK, map[string]interface{}{
			"logs": logs,
			"pagination": map[string]interface{}{
				"page":  page,
				"limit": limit,
				"total": total,
			},
		}, "観察対象に一致する行動記録を取得しました")
	}
}

func (h *TargetHandler) getTargetByID(w http.ResponseWriter, r *http.Request, id int) {
	userID, ok := currentUserID(r)
	if !ok {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "認証が必要です", nil)
		return
	}

	target, err := h.Usecase.GetTargetByID(r.Context(), id, userID)
	if err != nil {
		JSONError(w, http.StatusNotFound, "NOT_FOUND", "指定した観察対象が見つかりません", nil)
		return
	}

	JSONSuccess(w, http.StatusOK, target, "観察対象を取得しました")
}

func (h *TargetHandler) updateTarget(w http.ResponseWriter, r *http.Request, id int) {
	userID, ok := currentUserID(r)
	if !ok {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "認証が必要です", nil)
		return
	}

	var req struct {
		Name          string  `json:"name"`
		Description   string  `json:"description"`
		MatchMode     string  `json:"match_mode"`
		QueryText     string  `json:"query_text"`
		TagIDs        []int   `json:"tag_ids"`
		AnyTagIDs     []int   `json:"any_tag_ids"`
		AnyTagGroups  [][]int `json:"any_tag_groups"`
		ExcludeTagIDs []int   `json:"exclude_tag_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_JSON", "JSONの解析に失敗しました", err.Error())
		return
	}

	username, _ := r.Context().Value("username").(string)
	if username == "" {
		username = strconv.Itoa(userID)
	}

	target, err := h.Usecase.UpdateTarget(
		r.Context(),
		userID,
		id,
		req.Name,
		req.Description,
		username,
		req.MatchMode,
		req.QueryText,
		req.TagIDs,
		req.AnyTagIDs,
		req.ExcludeTagIDs,
		req.AnyTagGroups,
	)
	if err != nil {
		switch {
		case errors.Is(err, usecases.ErrTargetNameRequired),
			errors.Is(err, usecases.ErrTargetNameTooLong),
			errors.Is(err, usecases.ErrTargetTagIDsRequired),
			errors.Is(err, usecases.ErrTargetTagIDInvalid),
			errors.Is(err, usecases.ErrTargetMatchModeInvalid):
			JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			return
		case errors.Is(err, usecases.ErrTargetNameDuplicate):
			JSONError(w, http.StatusConflict, "CONFLICT", err.Error(), nil)
			return
		case errors.Is(err, usecases.ErrTargetTagNotFound):
			JSONError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
			return
		default:
			JSONError(w, http.StatusNotFound, "NOT_FOUND", "指定した観察対象が見つかりません", nil)
			return
		}
	}

	JSONSuccess(w, http.StatusOK, target, "観察対象を更新しました")
}

func (h *TargetHandler) deleteTarget(w http.ResponseWriter, r *http.Request, id int) {
	userID, ok := currentUserID(r)
	if !ok {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "認証が必要です", nil)
		return
	}

	if err := h.Usecase.DeleteTarget(r.Context(), userID, id); err != nil {
		JSONError(w, http.StatusNotFound, "NOT_FOUND", "指定した観察対象が見つかりません", nil)
		return
	}
	JSONSuccess(w, http.StatusOK, nil, "観察対象を削除しました")
}
