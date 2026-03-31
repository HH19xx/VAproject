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

type TargetHandler struct {
	Usecase *inferenceapp.TargetService
}

func (h *TargetHandler) ListTargetsHandler() http.HandlerFunc {
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
			sharedhttp.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch targets", err.Error())
			return
		}

		sharedhttp.JSONSuccess(w, http.StatusOK, map[string]interface{}{
			"targets": targets,
			"pagination": map[string]interface{}{
				"page":  page,
				"limit": limit,
				"total": total,
			},
		}, "targets fetched")
	}
}

func (h *TargetHandler) CreateTargetHandler() http.HandlerFunc {
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
			sharedhttp.JSONError(w, http.StatusBadRequest, "INVALID_JSON", "invalid json payload", err.Error())
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
			case errors.Is(err, inferenceapp.ErrTargetNameRequired),
				errors.Is(err, inferenceapp.ErrTargetNameTooLong),
				errors.Is(err, inferenceapp.ErrTargetTagIDsRequired),
				errors.Is(err, inferenceapp.ErrTargetTagIDInvalid),
				errors.Is(err, inferenceapp.ErrTargetMatchModeInvalid):
				sharedhttp.JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			case errors.Is(err, inferenceapp.ErrTargetNameDuplicate):
				sharedhttp.JSONError(w, http.StatusConflict, "CONFLICT", err.Error(), nil)
			case errors.Is(err, inferenceapp.ErrTargetTagNotFound):
				sharedhttp.JSONError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
			default:
				sharedhttp.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create target", err.Error())
			}
			return
		}

		sharedhttp.JSONSuccess(w, http.StatusCreated, target, "target created")
	}
}

func (h *TargetHandler) TargetDetailHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/targets/")
		idStr := strings.Split(path, "/")[0]
		id, err := strconv.Atoi(idStr)
		if err != nil || id < 1 {
			sharedhttp.JSONError(w, http.StatusBadRequest, "INVALID_ID", "invalid id", nil)
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
			sharedhttp.JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET/PUT/DELETE only", nil)
		}
	}
}

func (h *TargetHandler) ListActionLogsByTargetHandler() http.HandlerFunc {
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

		path := strings.TrimPrefix(r.URL.Path, "/api/v1/targets/")
		parts := strings.Split(path, "/")
		if len(parts) < 2 || parts[1] != "action_logs" {
			sharedhttp.JSONError(w, http.StatusBadRequest, "INVALID_PATH", "invalid path", nil)
			return
		}
		targetID, err := strconv.Atoi(parts[0])
		if err != nil || targetID < 1 {
			sharedhttp.JSONError(w, http.StatusBadRequest, "INVALID_ID", "invalid id", nil)
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
			sharedhttp.JSONError(w, http.StatusNotFound, "NOT_FOUND", "target or action logs not found", nil)
			return
		}

		sharedhttp.JSONSuccess(w, http.StatusOK, map[string]interface{}{
			"logs": logs,
			"pagination": map[string]interface{}{
				"page":  page,
				"limit": limit,
				"total": total,
			},
		}, "action logs by target fetched")
	}
}

func (h *TargetHandler) getTargetByID(w http.ResponseWriter, r *http.Request, id int) {
	userID, ok := currentUserID(r)
	if !ok {
		sharedhttp.JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}
	target, err := h.Usecase.GetTargetByID(r.Context(), id, userID)
	if err != nil {
		sharedhttp.JSONError(w, http.StatusNotFound, "NOT_FOUND", "target not found", nil)
		return
	}
	sharedhttp.JSONSuccess(w, http.StatusOK, target, "target fetched")
}

func (h *TargetHandler) updateTarget(w http.ResponseWriter, r *http.Request, id int) {
	userID, ok := currentUserID(r)
	if !ok {
		sharedhttp.JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
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
		sharedhttp.JSONError(w, http.StatusBadRequest, "INVALID_JSON", "invalid json payload", err.Error())
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
		case errors.Is(err, inferenceapp.ErrTargetNameRequired),
			errors.Is(err, inferenceapp.ErrTargetNameTooLong),
			errors.Is(err, inferenceapp.ErrTargetTagIDsRequired),
			errors.Is(err, inferenceapp.ErrTargetTagIDInvalid),
			errors.Is(err, inferenceapp.ErrTargetMatchModeInvalid):
			sharedhttp.JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		case errors.Is(err, inferenceapp.ErrTargetNameDuplicate):
			sharedhttp.JSONError(w, http.StatusConflict, "CONFLICT", err.Error(), nil)
		case errors.Is(err, inferenceapp.ErrTargetTagNotFound):
			sharedhttp.JSONError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
		default:
			sharedhttp.JSONError(w, http.StatusNotFound, "NOT_FOUND", "target not found", nil)
		}
		return
	}

	sharedhttp.JSONSuccess(w, http.StatusOK, target, "target updated")
}

func (h *TargetHandler) deleteTarget(w http.ResponseWriter, r *http.Request, id int) {
	userID, ok := currentUserID(r)
	if !ok {
		sharedhttp.JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}
	if err := h.Usecase.DeleteTarget(r.Context(), userID, id); err != nil {
		sharedhttp.JSONError(w, http.StatusNotFound, "NOT_FOUND", "target not found", nil)
		return
	}
	sharedhttp.JSONSuccess(w, http.StatusOK, nil, "target deleted")
}
