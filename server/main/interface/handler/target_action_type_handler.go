package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"server/main/usecases"
)

type TargetActionTypeHandler struct {
	Usecase *usecases.TargetActionTypeUsecase
}

// GET /api/v1/targets/{id}/action-types - 対象に紐づく行動種別一覧
func (h *TargetActionTypeHandler) ListByTargetHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GETのみ許可されています", nil)
			return
		}

		// URLからtarget_idを取得
		path := r.URL.Path
		parts := strings.Split(path, "/")
		// /api/v1/targets/{id}/action-types -> parts[4] = id
		if len(parts) < 5 {
			JSONError(w, http.StatusBadRequest, "INVALID_PATH", "パスが不正です", nil)
			return
		}
		targetID, err := strconv.Atoi(parts[4])
		if err != nil || targetID <= 0 {
			JSONError(w, http.StatusBadRequest, "INVALID_ID", "対象IDが不正です", nil)
			return
		}

		links, err := h.Usecase.GetActionTypesByTarget(r.Context(), targetID)
		if err != nil {
			JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "紐づけの取得に失敗しました", err.Error())
			return
		}

		JSONSuccess(w, http.StatusOK, map[string]interface{}{
			"links": links,
		}, "紐づけ一覧を取得しました")
	}
}

// POST /api/v1/targets/{id}/action-types - 対象に行動種別を紐づけ
func (h *TargetActionTypeHandler) LinkHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "POSTのみ許可されています", nil)
			return
		}

		path := r.URL.Path
		parts := strings.Split(path, "/")
		if len(parts) < 5 {
			JSONError(w, http.StatusBadRequest, "INVALID_PATH", "パスが不正です", nil)
			return
		}
		targetID, err := strconv.Atoi(parts[4])
		if err != nil || targetID <= 0 {
			JSONError(w, http.StatusBadRequest, "INVALID_ID", "対象IDが不正です", nil)
			return
		}

		var req struct {
			ActionTypeID int `json:"action_type_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			JSONError(w, http.StatusBadRequest, "INVALID_JSON", "JSONの解析に失敗しました", err.Error())
			return
		}

		username, _ := r.Context().Value("username").(string)
		if username == "" {
			username = "system"
		}

		link, err := h.Usecase.Link(r.Context(), targetID, req.ActionTypeID, username)
		if err != nil {
			switch err {
			case usecases.ErrTargetActionTypeAlreadyExists:
				JSONError(w, http.StatusConflict, "CONFLICT", err.Error(), nil)
			case usecases.ErrInvalidTargetID, usecases.ErrInvalidActionTypeID:
				JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			default:
				JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "紐づけの作成に失敗しました", err.Error())
			}
			return
		}

		JSONSuccess(w, http.StatusCreated, link, "紐づけを作成しました")
	}
}

// DELETE /api/v1/targets/{id}/action-types/{action_type_id} - 紐づけ解除
func (h *TargetActionTypeHandler) UnlinkHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "DELETEのみ許可されています", nil)
			return
		}

		path := r.URL.Path
		parts := strings.Split(path, "/")
		// /api/v1/targets/{id}/action-types/{action_type_id}
		if len(parts) < 7 {
			JSONError(w, http.StatusBadRequest, "INVALID_PATH", "パスが不正です", nil)
			return
		}
		targetID, err := strconv.Atoi(parts[4])
		if err != nil || targetID <= 0 {
			JSONError(w, http.StatusBadRequest, "INVALID_ID", "対象IDが不正です", nil)
			return
		}
		actionTypeID, err := strconv.Atoi(parts[6])
		if err != nil || actionTypeID <= 0 {
			JSONError(w, http.StatusBadRequest, "INVALID_ID", "行動種別IDが不正です", nil)
			return
		}

		if err := h.Usecase.Unlink(r.Context(), targetID, actionTypeID); err != nil {
			JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "紐づけの解除に失敗しました", err.Error())
			return
		}

		JSONSuccess(w, http.StatusOK, nil, "紐づけを解除しました")
	}
}
