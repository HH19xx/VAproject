package httpadapter

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	inferenceapp "server/internal/modules/inference/app"
	sharedhttp "server/internal/shared/http"
)

type PrototypeHandler struct {
	Usecase *inferenceapp.PrototypeService
}

func (h *PrototypeHandler) ListHandler() http.HandlerFunc {
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
		prototypes, err := h.Usecase.GetAll(r.Context(), userID)
		if err != nil {
			sharedhttp.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch prototypes", err.Error())
			return
		}
		sharedhttp.JSONSuccess(w, http.StatusOK, map[string]interface{}{"prototypes": prototypes}, "prototypes fetched")
	}
}

func (h *PrototypeHandler) CreateHandler() http.HandlerFunc {
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
			Name              string `json:"name"`
			Description       string `json:"description"`
			ParentPrototypeID *int   `json:"parent_prototype_id"`
			TagIDs            []int  `json:"tag_ids"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sharedhttp.JSONError(w, http.StatusBadRequest, "INVALID_JSON", "invalid json payload", err.Error())
			return
		}

		username, _ := r.Context().Value("username").(string)
		if username == "" {
			username = "system"
		}

		prototype, err := h.Usecase.Create(r.Context(), userID, req.Name, req.Description, username, req.ParentPrototypeID, req.TagIDs)
		if err != nil {
			switch {
			case errors.Is(err, inferenceapp.ErrPrototypeNameRequired),
				errors.Is(err, inferenceapp.ErrPrototypeNameTooLong),
				errors.Is(err, inferenceapp.ErrPrototypeTagIDsRequired),
				errors.Is(err, inferenceapp.ErrPrototypeTagIDInvalid),
				errors.Is(err, inferenceapp.ErrPrototypeParentNotFound),
				errors.Is(err, inferenceapp.ErrPrototypeParentNotOwned):
				sharedhttp.JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			case errors.Is(err, inferenceapp.ErrPrototypeNameDuplicate):
				sharedhttp.JSONError(w, http.StatusConflict, "CONFLICT", err.Error(), nil)
			case errors.Is(err, inferenceapp.ErrPrototypeTagNotFound):
				sharedhttp.JSONError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
			default:
				sharedhttp.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create prototype", err.Error())
			}
			return
		}
		sharedhttp.JSONSuccess(w, http.StatusCreated, prototype, "prototype created")
	}
}

func (h *PrototypeHandler) DetailHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/prototypes/")
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

func (h *PrototypeHandler) getByID(w http.ResponseWriter, r *http.Request, id int) {
	userID, ok := currentUserID(r)
	if !ok {
		sharedhttp.JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}
	prototype, err := h.Usecase.GetByID(r.Context(), id, userID)
	if err != nil {
		sharedhttp.JSONError(w, http.StatusNotFound, "NOT_FOUND", "prototype not found", nil)
		return
	}
	sharedhttp.JSONSuccess(w, http.StatusOK, prototype, "prototype fetched")
}

func (h *PrototypeHandler) update(w http.ResponseWriter, r *http.Request, id int) {
	userID, ok := currentUserID(r)
	if !ok {
		sharedhttp.JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}

	var req struct {
		Name              string `json:"name"`
		Description       string `json:"description"`
		ParentPrototypeID *int   `json:"parent_prototype_id"`
		TagIDs            []int  `json:"tag_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedhttp.JSONError(w, http.StatusBadRequest, "INVALID_JSON", "invalid json payload", err.Error())
		return
	}

	username, _ := r.Context().Value("username").(string)
	if username == "" {
		username = "system"
	}
	prototype, err := h.Usecase.Update(r.Context(), userID, id, req.Name, req.Description, username, req.ParentPrototypeID, req.TagIDs)
	if err != nil {
		switch {
		case errors.Is(err, inferenceapp.ErrPrototypeNameRequired),
			errors.Is(err, inferenceapp.ErrPrototypeNameTooLong),
			errors.Is(err, inferenceapp.ErrPrototypeTagIDsRequired),
			errors.Is(err, inferenceapp.ErrPrototypeTagIDInvalid),
			errors.Is(err, inferenceapp.ErrPrototypeParentNotFound),
			errors.Is(err, inferenceapp.ErrPrototypeParentNotOwned),
			errors.Is(err, inferenceapp.ErrPrototypeParentSelfNotAllow):
			sharedhttp.JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		case errors.Is(err, inferenceapp.ErrPrototypeNameDuplicate):
			sharedhttp.JSONError(w, http.StatusConflict, "CONFLICT", err.Error(), nil)
		case errors.Is(err, inferenceapp.ErrPrototypeTagNotFound):
			sharedhttp.JSONError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
		default:
			sharedhttp.JSONError(w, http.StatusNotFound, "NOT_FOUND", "prototype not found", nil)
		}
		return
	}
	sharedhttp.JSONSuccess(w, http.StatusOK, prototype, "prototype updated")
}

func (h *PrototypeHandler) delete(w http.ResponseWriter, r *http.Request, id int) {
	userID, ok := currentUserID(r)
	if !ok {
		sharedhttp.JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}
	if err := h.Usecase.Delete(r.Context(), userID, id); err != nil {
		sharedhttp.JSONError(w, http.StatusNotFound, "NOT_FOUND", "prototype not found", nil)
		return
	}
	sharedhttp.JSONSuccess(w, http.StatusOK, nil, "prototype deleted")
}
