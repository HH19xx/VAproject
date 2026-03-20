package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"server/main/usecases"
)

type PrototypeHandler struct {
	Usecase *usecases.PrototypeUsecase
}

// GET /api/v1/prototypes
func (h *PrototypeHandler) ListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GETメソッドのみ許可されています", nil)
			return
		}
		userID, ok := currentUserID(r)
		if !ok {
			JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "認証が必要です", nil)
			return
		}

		prototypes, err := h.Usecase.GetAll(r.Context(), userID)
		if err != nil {
			JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "プロトタイプ一覧の取得に失敗しました", err.Error())
			return
		}
		JSONSuccess(w, http.StatusOK, map[string]interface{}{
			"prototypes": prototypes,
		}, "プロトタイプ一覧を取得しました")
	}
}

// POST /api/v1/prototypes
func (h *PrototypeHandler) CreateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "POSTメソッドのみ許可されています", nil)
			return
		}
		userID, ok := currentUserID(r)
		if !ok {
			JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "認証が必要です", nil)
			return
		}

		var req struct {
			Name              string `json:"name"`
			Description       string `json:"description"`
			ParentPrototypeID *int   `json:"parent_prototype_id"`
			TagIDs            []int  `json:"tag_ids"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			JSONError(w, http.StatusBadRequest, "INVALID_JSON", "JSONの解析に失敗しました", err.Error())
			return
		}

		username, _ := r.Context().Value("username").(string)
		if username == "" {
			username = "system"
		}

		prototype, err := h.Usecase.Create(r.Context(), userID, req.Name, req.Description, username, req.ParentPrototypeID, req.TagIDs)
		if err != nil {
			switch {
			case errors.Is(err, usecases.ErrPrototypeNameRequired),
				errors.Is(err, usecases.ErrPrototypeNameTooLong),
				errors.Is(err, usecases.ErrPrototypeTagIDsRequired),
				errors.Is(err, usecases.ErrPrototypeTagIDInvalid),
				errors.Is(err, usecases.ErrPrototypeParentNotFound),
				errors.Is(err, usecases.ErrPrototypeParentNotOwned):
				JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
				return
			case errors.Is(err, usecases.ErrPrototypeNameDuplicate):
				JSONError(w, http.StatusConflict, "CONFLICT", err.Error(), nil)
				return
			case errors.Is(err, usecases.ErrPrototypeTagNotFound):
				JSONError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
				return
			default:
				JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "プロトタイプの作成に失敗しました", err.Error())
				return
			}
		}
		JSONSuccess(w, http.StatusCreated, prototype, "プロトタイプを作成しました")
	}
}

// GET/PUT/DELETE /api/v1/prototypes/{id}
func (h *PrototypeHandler) DetailHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/prototypes/")
		idStr := strings.Split(path, "/")[0]
		id, err := strconv.Atoi(idStr)
		if err != nil || id < 1 {
			JSONError(w, http.StatusBadRequest, "INVALID_ID", "IDが不正です", nil)
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
			JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET/PUT/DELETEメソッドのみ許可されています", nil)
		}
	}
}

func (h *PrototypeHandler) getByID(w http.ResponseWriter, r *http.Request, id int) {
	userID, ok := currentUserID(r)
	if !ok {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "認証が必要です", nil)
		return
	}

	prototype, err := h.Usecase.GetByID(r.Context(), id, userID)
	if err != nil {
		JSONError(w, http.StatusNotFound, "NOT_FOUND", "プロトタイプが見つかりません", nil)
		return
	}
	JSONSuccess(w, http.StatusOK, prototype, "プロトタイプを取得しました")
}

func (h *PrototypeHandler) update(w http.ResponseWriter, r *http.Request, id int) {
	userID, ok := currentUserID(r)
	if !ok {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "認証が必要です", nil)
		return
	}

	var req struct {
		Name              string `json:"name"`
		Description       string `json:"description"`
		ParentPrototypeID *int   `json:"parent_prototype_id"`
		TagIDs            []int  `json:"tag_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_JSON", "JSONの解析に失敗しました", err.Error())
		return
	}

	username, _ := r.Context().Value("username").(string)
	if username == "" {
		username = "system"
	}
	prototype, err := h.Usecase.Update(r.Context(), userID, id, req.Name, req.Description, username, req.ParentPrototypeID, req.TagIDs)
	if err != nil {
		switch {
		case errors.Is(err, usecases.ErrPrototypeNameRequired),
			errors.Is(err, usecases.ErrPrototypeNameTooLong),
			errors.Is(err, usecases.ErrPrototypeTagIDsRequired),
			errors.Is(err, usecases.ErrPrototypeTagIDInvalid),
			errors.Is(err, usecases.ErrPrototypeParentNotFound),
			errors.Is(err, usecases.ErrPrototypeParentNotOwned),
			errors.Is(err, usecases.ErrPrototypeParentSelfNotAllow):
			JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			return
		case errors.Is(err, usecases.ErrPrototypeNameDuplicate):
			JSONError(w, http.StatusConflict, "CONFLICT", err.Error(), nil)
			return
		case errors.Is(err, usecases.ErrPrototypeTagNotFound):
			JSONError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
			return
		default:
			JSONError(w, http.StatusNotFound, "NOT_FOUND", "プロトタイプが見つかりません", nil)
			return
		}
	}
	JSONSuccess(w, http.StatusOK, prototype, "プロトタイプを更新しました")
}

func (h *PrototypeHandler) delete(w http.ResponseWriter, r *http.Request, id int) {
	userID, ok := currentUserID(r)
	if !ok {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "認証が必要です", nil)
		return
	}

	if err := h.Usecase.Delete(r.Context(), userID, id); err != nil {
		JSONError(w, http.StatusNotFound, "NOT_FOUND", "プロトタイプが見つかりません", nil)
		return
	}
	JSONSuccess(w, http.StatusOK, nil, "プロトタイプを削除しました")
}
