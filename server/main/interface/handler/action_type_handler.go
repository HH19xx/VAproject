package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"server/main/usecases"
)

type ActionTypeHandler struct {
	Usecase *usecases.ActionTypeUsecase
}

// GET /api/v1/action_types
func (h *ActionTypeHandler) ListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GETメソッドのみ許可されています", nil)
			return
		}
		types, err := h.Usecase.GetAll(r.Context())
		if err != nil {
			JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "行動種別の取得に失敗しました", err.Error())
			return
		}
		JSONSuccess(w, http.StatusOK, map[string]interface{}{
			"action_types": types,
		}, "行動種別を取得しました")
	}
}

// POST /api/v1/action_types
func (h *ActionTypeHandler) CreateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "POSTメソッドのみ許可されています", nil)
			return
		}
		var req struct {
			ActionName  string `json:"action_name"`
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
		at, err := h.Usecase.Create(r.Context(), req.ActionName, req.Description, username)
		if err != nil {
			if err == usecases.ErrActionTypeNameRequired || err == usecases.ErrActionTypeNameTooLong {
				JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
				return
			}
			JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "行動種別の作成に失敗しました", err.Error())
			return
		}
		JSONSuccess(w, http.StatusCreated, at, "行動種別を作成しました")
	}
}

// GET/PUT/DELETE /api/v1/action_types/{id}
func (h *ActionTypeHandler) DetailHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/action_types/")
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

func (h *ActionTypeHandler) getByID(w http.ResponseWriter, r *http.Request, id int) {
	at, err := h.Usecase.GetByID(r.Context(), id)
	if err != nil {
		JSONError(w, http.StatusNotFound, "NOT_FOUND", "行動種別が見つかりません", nil)
		return
	}
	JSONSuccess(w, http.StatusOK, at, "行動種別を取得しました")
}

func (h *ActionTypeHandler) update(w http.ResponseWriter, r *http.Request, id int) {
	var req struct {
		ActionName  string `json:"action_name"`
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
	at, err := h.Usecase.Update(r.Context(), id, req.ActionName, req.Description, username)
	if err != nil {
		if err == usecases.ErrActionTypeNameRequired || err == usecases.ErrActionTypeNameTooLong {
			JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			return
		}
		JSONError(w, http.StatusNotFound, "NOT_FOUND", "行動種別が見つかりません", nil)
		return
	}
	JSONSuccess(w, http.StatusOK, at, "行動種別を更新しました")
}

func (h *ActionTypeHandler) delete(w http.ResponseWriter, r *http.Request, id int) {
	if err := h.Usecase.Delete(r.Context(), id); err != nil {
		JSONError(w, http.StatusNotFound, "NOT_FOUND", "行動種別が見つかりません", nil)
		return
	}
	JSONSuccess(w, http.StatusOK, nil, "行動種別を削除しました")
}
