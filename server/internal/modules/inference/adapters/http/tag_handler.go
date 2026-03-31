package httpadapter

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	inferenceapp "server/internal/modules/inference/app"
	sharedhttp "server/internal/shared/http"
)

type TagHandler struct {
	Usecase *inferenceapp.TagService
}

func (h *TagHandler) ListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			sharedhttp.JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET method only", nil)
			return
		}
		userID, ok := currentUserID(r)
		if !ok {
			sharedhttp.JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
			return
		}
		tags, err := h.Usecase.GetAll(r.Context(), userID)
		if err != nil {
			sharedhttp.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch tags", err.Error())
			return
		}
		sharedhttp.JSONSuccess(w, http.StatusOK, map[string]interface{}{"tags": tags}, "tags fetched")
	}
}

func (h *TagHandler) CreateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			sharedhttp.JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "POST method only", nil)
			return
		}
		userID, ok := currentUserID(r)
		if !ok {
			sharedhttp.JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
			return
		}

		var req struct {
			Name        string `json:"name"`
			Group       string `json:"group"`
			Description string `json:"description"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sharedhttp.JSONError(w, http.StatusBadRequest, "INVALID_JSON", "invalid json payload", err.Error())
			return
		}

		username, _ := r.Context().Value("username").(string)
		if username == "" {
			username = "system"
		}

		tag, err := h.Usecase.Create(r.Context(), userID, req.Name, req.Group, req.Description, username)
		if err != nil {
			switch err {
			case inferenceapp.ErrTagNameRequired, inferenceapp.ErrTagNameTooLong:
				sharedhttp.JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			case inferenceapp.ErrTagNameDuplicate:
				sharedhttp.JSONError(w, http.StatusConflict, "CONFLICT", err.Error(), nil)
			default:
				sharedhttp.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create tag", err.Error())
			}
			return
		}
		sharedhttp.JSONSuccess(w, http.StatusCreated, tag, "tag created")
	}
}

func (h *TagHandler) DetailHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/tags/")
		idStr := strings.Split(path, "/")[0]
		id, err := strconv.Atoi(idStr)
		if err != nil || id < 1 {
			sharedhttp.JSONError(w, http.StatusBadRequest, "INVALID_ID", "invalid id", nil)
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
			sharedhttp.JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET/PUT/DELETE only", nil)
		}
	}
}

func (h *TagHandler) getByID(w http.ResponseWriter, r *http.Request, id int) {
	userID, ok := currentUserID(r)
	if !ok {
		sharedhttp.JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}
	tag, err := h.Usecase.GetByID(r.Context(), id, userID)
	if err != nil {
		sharedhttp.JSONError(w, http.StatusNotFound, "NOT_FOUND", "tag not found", nil)
		return
	}
	sharedhttp.JSONSuccess(w, http.StatusOK, tag, "tag fetched")
}

func (h *TagHandler) update(w http.ResponseWriter, r *http.Request, id int) {
	userID, ok := currentUserID(r)
	if !ok {
		sharedhttp.JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}

	var req struct {
		Name        string `json:"name"`
		Group       string `json:"group"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedhttp.JSONError(w, http.StatusBadRequest, "INVALID_JSON", "invalid json payload", err.Error())
		return
	}
	username, _ := r.Context().Value("username").(string)
	if username == "" {
		username = "system"
	}

	tag, err := h.Usecase.Update(r.Context(), userID, id, req.Name, req.Group, req.Description, username)
	if err != nil {
		switch err {
		case inferenceapp.ErrTagNameRequired, inferenceapp.ErrTagNameTooLong:
			sharedhttp.JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		case inferenceapp.ErrTagNameDuplicate:
			sharedhttp.JSONError(w, http.StatusConflict, "CONFLICT", err.Error(), nil)
		default:
			sharedhttp.JSONError(w, http.StatusNotFound, "NOT_FOUND", "tag not found", nil)
		}
		return
	}
	sharedhttp.JSONSuccess(w, http.StatusOK, tag, "tag updated")
}

func (h *TagHandler) delete(w http.ResponseWriter, r *http.Request, id int) {
	userID, ok := currentUserID(r)
	if !ok {
		sharedhttp.JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}
	if err := h.Usecase.Delete(r.Context(), userID, id); err != nil {
		sharedhttp.JSONError(w, http.StatusNotFound, "NOT_FOUND", "tag not found", nil)
		return
	}
	sharedhttp.JSONSuccess(w, http.StatusOK, nil, "tag deleted")
}
