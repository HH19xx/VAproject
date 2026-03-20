package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"server/main/domain"
	"server/main/usecases"
)

type ActionLogHandler struct {
	Usecase *usecases.ActionLogUsecase
}

// GET /api/v1/action_logs?tag_ids=1,2,3&any_tag_ids=4,5&any_tag_groups=4+5,6+7&exclude_tag_ids=8&page=&limit=&sort=&order=&from=&to=
func (h *ActionLogHandler) ListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET method only", nil)
			return
		}
		userID, ok := currentUserID(r)
		if !ok {
			JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
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

		tagIDs, err := parseTagIDsCSV(r.URL.Query().Get("tag_ids"))
		if err != nil {
			JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", "tag_ids must be comma separated positive integers", nil)
			return
		}

		listOptions, err := parseActionLogListOptions(r)
		if err != nil {
			JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			return
		}

		anyTagIDs, err := parseTagIDsCSV(r.URL.Query().Get("any_tag_ids"))
		if err != nil {
			JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", "any_tag_ids must be comma separated positive integers", nil)
			return
		}
		anyTagGroups, err := parseTagIDGroupsCSV(r.URL.Query().Get("any_tag_groups"))
		if err != nil {
			JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", "any_tag_groups must be comma separated groups of '+' joined positive integers", nil)
			return
		}
		excludeTagIDs, err := parseTagIDsCSV(r.URL.Query().Get("exclude_tag_ids"))
		if err != nil {
			JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", "exclude_tag_ids must be comma separated positive integers", nil)
			return
		}
		listOptions.AnyTagIDs = anyTagIDs
		listOptions.AnyTagGroups = anyTagGroups
		listOptions.ExcludeTagIDs = excludeTagIDs

		logs, total, meta, err := h.Usecase.GetActionLogs(r.Context(), userID, tagIDs, page, limit, listOptions)
		if err != nil {
			switch {
			case errors.Is(err, usecases.ErrActionLogSortInvalid), errors.Is(err, usecases.ErrActionLogOrderInvalid):
				JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
				return
			default:
				JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch action logs", err.Error())
				return
			}
		}
		JSONSuccess(w, http.StatusOK, map[string]interface{}{
			"logs": logs,
			"pagination": map[string]interface{}{
				"page":  page,
				"limit": limit,
				"total": total,
			},
			"meta": meta,
		}, "action logs fetched")
	}
}

// POST /api/v1/action_logs
func (h *ActionLogHandler) CreateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "POST method only", nil)
			return
		}
		userID, ok := currentUserID(r)
		if !ok {
			JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
			return
		}
		var req struct {
			Title             string `json:"title"`
			ParentPrototypeID *int   `json:"parent_prototype_id"`
			PrototypeID       *int   `json:"prototype_id"` // backward compatibility
			OccurredAt        string `json:"occurred_at"`
			Timestamp         string `json:"timestamp"` // backward compatibility
			Notes             string `json:"notes"`
			TagIDs            []int  `json:"tag_ids"`
			Attributes        []struct {
				Key         string   `json:"key"`
				ValueNumber *float64 `json:"value_number"`
			} `json:"attributes"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			JSONError(w, http.StatusBadRequest, "INVALID_JSON", "invalid json payload", err.Error())
			return
		}

		parentPrototypeID := req.ParentPrototypeID
		if parentPrototypeID == nil {
			parentPrototypeID = req.PrototypeID
		}

		occurredAtRaw := req.OccurredAt
		if occurredAtRaw == "" {
			occurredAtRaw = req.Timestamp
		}
		occurredAt, err := time.Parse(time.RFC3339, occurredAtRaw)
		if err != nil {
			JSONError(w, http.StatusBadRequest, "INVALID_TIMESTAMP", "occurred_at must be RFC3339", nil)
			return
		}
		username, _ := r.Context().Value("username").(string)
		if username == "" {
			username = "system"
		}
		attributes := make([]domain.ActionLogAttribute, 0, len(req.Attributes))
		for _, attr := range req.Attributes {
			attributes = append(attributes, domain.ActionLogAttribute{
				Key:         attr.Key,
				ValueNumber: attr.ValueNumber,
			})
		}
		log, err := h.Usecase.CreateActionLog(
			r.Context(),
			userID,
			req.Title,
			parentPrototypeID,
			occurredAt,
			req.Notes,
			username,
			req.TagIDs,
			attributes,
		)
		if err != nil {
			switch {
			case errors.Is(err, usecases.ErrActionLogTitleRequired),
				errors.Is(err, usecases.ErrActionLogTitleTooLong),
				errors.Is(err, usecases.ErrActionLogOccurredAtRequired),
				errors.Is(err, usecases.ErrActionLogTagIDsRequired),
				errors.Is(err, usecases.ErrActionLogTagIDInvalid),
				errors.Is(err, usecases.ErrActionLogAttributeKeyRequired),
				errors.Is(err, usecases.ErrActionLogAttributeKeyTooLong),
				errors.Is(err, usecases.ErrActionLogAttributeValueRequired),
				errors.Is(err, usecases.ErrActionLogAttributeValueInvalid),
				errors.Is(err, usecases.ErrActionLogParentPrototypeIDInvalid),
				errors.Is(err, usecases.ErrActionLogParentPrototypeSelf):
				JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
				return
			case errors.Is(err, usecases.ErrActionLogTagNotFound),
				errors.Is(err, usecases.ErrActionLogParentPrototypeNotFound):
				JSONError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
				return
			default:
				JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create action log", err.Error())
				return
			}
		}
		JSONSuccess(w, http.StatusCreated, log, "action log created")
	}
}

// GET/PUT/DELETE /api/v1/action_logs/{id}
func (h *ActionLogHandler) DetailHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/action_logs/")
		idStr := strings.Split(path, "/")[0]
		id, err := strconv.Atoi(idStr)
		if err != nil || id < 1 {
			JSONError(w, http.StatusBadRequest, "INVALID_ID", "invalid id", nil)
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
			JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET/PUT/DELETE only", nil)
		}
	}
}

func (h *ActionLogHandler) getByID(w http.ResponseWriter, r *http.Request, id int) {
	userID, ok := currentUserID(r)
	if !ok {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}

	log, err := h.Usecase.GetActionLogByID(r.Context(), userID, id)
	if err != nil {
		JSONError(w, http.StatusNotFound, "NOT_FOUND", "action log not found", nil)
		return
	}
	JSONSuccess(w, http.StatusOK, log, "action log fetched")
}

func (h *ActionLogHandler) update(w http.ResponseWriter, r *http.Request, id int) {
	userID, ok := currentUserID(r)
	if !ok {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}

	var req struct {
		Title             string `json:"title"`
		ParentPrototypeID *int   `json:"parent_prototype_id"`
		PrototypeID       *int   `json:"prototype_id"` // backward compatibility
		OccurredAt        string `json:"occurred_at"`
		Timestamp         string `json:"timestamp"` // backward compatibility
		Notes             string `json:"notes"`
		TagIDs            []int  `json:"tag_ids"`
		Attributes        []struct {
			Key         string   `json:"key"`
			ValueNumber *float64 `json:"value_number"`
		} `json:"attributes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_JSON", "invalid json payload", err.Error())
		return
	}

	parentPrototypeID := req.ParentPrototypeID
	if parentPrototypeID == nil {
		parentPrototypeID = req.PrototypeID
	}

	occurredAtRaw := req.OccurredAt
	if occurredAtRaw == "" {
		occurredAtRaw = req.Timestamp
	}
	occurredAt, err := time.Parse(time.RFC3339, occurredAtRaw)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_TIMESTAMP", "occurred_at must be RFC3339", nil)
		return
	}
	username, _ := r.Context().Value("username").(string)
	if username == "" {
		username = "system"
	}
	attributes := make([]domain.ActionLogAttribute, 0, len(req.Attributes))
	for _, attr := range req.Attributes {
		attributes = append(attributes, domain.ActionLogAttribute{
			Key:         attr.Key,
			ValueNumber: attr.ValueNumber,
		})
	}
	log, err := h.Usecase.UpdateActionLog(
		r.Context(),
		userID,
		id,
		req.Title,
		parentPrototypeID,
		occurredAt,
		req.Notes,
		username,
		req.TagIDs,
		attributes,
	)
	if err != nil {
		switch {
		case errors.Is(err, usecases.ErrActionLogTitleRequired),
			errors.Is(err, usecases.ErrActionLogTitleTooLong),
			errors.Is(err, usecases.ErrActionLogOccurredAtRequired),
			errors.Is(err, usecases.ErrActionLogTagIDsRequired),
			errors.Is(err, usecases.ErrActionLogTagIDInvalid),
			errors.Is(err, usecases.ErrActionLogAttributeKeyRequired),
			errors.Is(err, usecases.ErrActionLogAttributeKeyTooLong),
			errors.Is(err, usecases.ErrActionLogAttributeValueRequired),
			errors.Is(err, usecases.ErrActionLogAttributeValueInvalid),
			errors.Is(err, usecases.ErrActionLogParentPrototypeIDInvalid),
			errors.Is(err, usecases.ErrActionLogParentPrototypeSelf):
			JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			return
		case errors.Is(err, usecases.ErrActionLogTagNotFound),
			errors.Is(err, usecases.ErrActionLogParentPrototypeNotFound):
			JSONError(w, http.StatusNotFound, "NOT_FOUND", err.Error(), nil)
			return
		default:
			JSONError(w, http.StatusNotFound, "NOT_FOUND", "action log not found", nil)
			return
		}
	}
	JSONSuccess(w, http.StatusOK, log, "action log updated")
}

func (h *ActionLogHandler) delete(w http.ResponseWriter, r *http.Request, id int) {
	userID, ok := currentUserID(r)
	if !ok {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
		return
	}

	if err := h.Usecase.DeleteActionLog(r.Context(), userID, id); err != nil {
		JSONError(w, http.StatusNotFound, "NOT_FOUND", "action log not found", nil)
		return
	}
	JSONSuccess(w, http.StatusOK, nil, "action log deleted")
}

func parseTagIDsCSV(raw string) ([]int, error) {
	if raw == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	tagIDs := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		id, err := strconv.Atoi(p)
		if err != nil || id <= 0 {
			return nil, errors.New("invalid tag id")
		}
		tagIDs = append(tagIDs, id)
	}
	return tagIDs, nil
}

func parseTagIDGroupsCSV(raw string) ([][]int, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	groupParts := strings.Split(raw, ",")
	groups := make([][]int, 0, len(groupParts))
	for _, groupRaw := range groupParts {
		groupRaw = strings.TrimSpace(groupRaw)
		if groupRaw == "" {
			continue
		}
		tagParts := strings.Split(groupRaw, "+")
		group := make([]int, 0, len(tagParts))
		for _, tagRaw := range tagParts {
			tagRaw = strings.TrimSpace(tagRaw)
			if tagRaw == "" {
				continue
			}
			id, err := strconv.Atoi(tagRaw)
			if err != nil || id <= 0 {
				return nil, errors.New("invalid tag id in group")
			}
			group = append(group, id)
		}
		if len(group) > 0 {
			groups = append(groups, group)
		}
	}
	return groups, nil
}

func parseActionLogListOptions(r *http.Request) (usecases.ActionLogListOptions, error) {
	var options usecases.ActionLogListOptions
	options.Sort = strings.TrimSpace(r.URL.Query().Get("sort"))
	options.Order = strings.TrimSpace(r.URL.Query().Get("order"))

	fromRaw := strings.TrimSpace(r.URL.Query().Get("from"))
	if fromRaw != "" {
		from, err := time.Parse(time.RFC3339, fromRaw)
		if err != nil {
			return options, errors.New("from must be RFC3339")
		}
		options.From = &from
	}

	toRaw := strings.TrimSpace(r.URL.Query().Get("to"))
	if toRaw != "" {
		to, err := time.Parse(time.RFC3339, toRaw)
		if err != nil {
			return options, errors.New("to must be RFC3339")
		}
		options.To = &to
	}
	return options, nil
}
