package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"server/main/usecases"
)

type TagHandler struct {
	Usecase *usecases.TagUsecase
}

// GET /api/v1/tags
func (h *TagHandler) ListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GETメソッドのみ許可されています", nil)
			return
		}
		userID, ok := currentUserID(r)
		if !ok {
			JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "認証情報が不正です", nil)
			return
		}

		tags, err := h.Usecase.GetAll(r.Context(), userID)
		if err != nil {
			JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "タグの取得に失敗しました", err.Error())
			return
		}
		JSONSuccess(w, http.StatusOK, map[string]interface{}{
			"tags": tags,
		}, "タグ一覧を取得しました")
	}
}

// POST /api/v1/tags
func (h *TagHandler) CreateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "POSTメソッドのみ許可されています", nil)
			return
		}
		userID, ok := currentUserID(r)
		if !ok {
			JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "認証情報が不正です", nil)
			return
		}

		var req struct {
			Name        string `json:"name"`
			Group       string `json:"group"`
			Description string `json:"description"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			JSONError(w, http.StatusBadRequest, "INVALID_JSON", "JSONの解析に失敗しました", err.Error())
			return
		}
		username, _ := r.Context().Value("username").(string)
		if username == "" {
			username = "system"
		}
		tag, err := h.Usecase.Create(r.Context(), userID, req.Name, req.Group, req.Description, username)
		if err != nil {
			switch err {
			case usecases.ErrTagNameRequired, usecases.ErrTagNameTooLong:
				JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
				return
			case usecases.ErrTagNameDuplicate:
				JSONError(w, http.StatusConflict, "CONFLICT", err.Error(), nil)
				return
			default:
				JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "タグの作成に失敗しました", err.Error())
				return
			}
		}
		JSONSuccess(w, http.StatusCreated, tag, "タグを作成しました")
	}
}

// GET/PUT/DELETE /api/v1/tags/{id}
func (h *TagHandler) DetailHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/tags/")
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

func (h *TagHandler) getByID(w http.ResponseWriter, r *http.Request, id int) {
	userID, ok := currentUserID(r)
	if !ok {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "認証情報が不正です", nil)
		return
	}

	tag, err := h.Usecase.GetByID(r.Context(), id, userID)
	if err != nil {
		JSONError(w, http.StatusNotFound, "NOT_FOUND", "タグが見つかりません", nil)
		return
	}
	JSONSuccess(w, http.StatusOK, tag, "タグを取得しました")
}

func (h *TagHandler) update(w http.ResponseWriter, r *http.Request, id int) {
	userID, ok := currentUserID(r)
	if !ok {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "認証情報が不正です", nil)
		return
	}

	var req struct {
		Name        string `json:"name"`
		Group       string `json:"group"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_JSON", "JSONの解析に失敗しました", err.Error())
		return
	}
	username, _ := r.Context().Value("username").(string)
	if username == "" {
		username = "system"
	}
	tag, err := h.Usecase.Update(r.Context(), userID, id, req.Name, req.Group, req.Description, username)
	if err != nil {
		switch err {
		case usecases.ErrTagNameRequired, usecases.ErrTagNameTooLong:
			JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			return
		case usecases.ErrTagNameDuplicate:
			JSONError(w, http.StatusConflict, "CONFLICT", err.Error(), nil)
			return
		default:
			JSONError(w, http.StatusNotFound, "NOT_FOUND", "タグが見つかりません", nil)
			return
		}
	}
	JSONSuccess(w, http.StatusOK, tag, "タグを更新しました")
}

func (h *TagHandler) delete(w http.ResponseWriter, r *http.Request, id int) {
	userID, ok := currentUserID(r)
	if !ok {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "認証情報が不正です", nil)
		return
	}

	if err := h.Usecase.Delete(r.Context(), userID, id); err != nil {
		JSONError(w, http.StatusNotFound, "NOT_FOUND", "タグが見つかりません", nil)
		return
	}
	JSONSuccess(w, http.StatusOK, nil, "タグを削除しました")
}
