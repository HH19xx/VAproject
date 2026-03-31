package httpadapter

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	deferenceapp "server/internal/modules/deference/app"
	inferenceapp "server/internal/modules/inference/app"
	externalsignalapp "server/internal/modules/externalsignal/app"
	sharedhttp "server/internal/shared/http"
)

type DistributionAnalysisHandler struct {
	Usecase *deferenceapp.DistributionService
}

func (h *DistributionAnalysisHandler) AnalyzeHandler() http.HandlerFunc {
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

		tagIDs, err := parseTagIDsCSV(r.URL.Query().Get("tag_ids"))
		if err != nil {
			sharedhttp.JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", "tag_ids must be comma separated positive integers", nil)
			return
		}
		anyTagIDs, err := parseTagIDsCSV(r.URL.Query().Get("any_tag_ids"))
		if err != nil {
			sharedhttp.JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", "any_tag_ids must be comma separated positive integers", nil)
			return
		}
		anyTagGroups, err := parseTagIDGroupsCSV(r.URL.Query().Get("any_tag_groups"))
		if err != nil {
			sharedhttp.JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", "any_tag_groups must be comma separated groups of '+' joined positive integers", nil)
			return
		}
		excludeTagIDs, err := parseTagIDsCSV(r.URL.Query().Get("exclude_tag_ids"))
		if err != nil {
			sharedhttp.JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", "exclude_tag_ids must be comma separated positive integers", nil)
			return
		}

		options, err := parseActionLogListOptions(r)
		if err != nil {
			sharedhttp.JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			return
		}
		options.AnyTagIDs = anyTagIDs
		options.AnyTagGroups = anyTagGroups
		options.ExcludeTagIDs = excludeTagIDs

		dataset := strings.TrimSpace(r.URL.Query().Get("dataset"))
		if dataset == "" {
			dataset = string(deferenceapp.DistributionDatasetActionLogs)
		}
		axis := strings.TrimSpace(r.URL.Query().Get("axis"))
		source := strings.TrimSpace(r.URL.Query().Get("source"))
		signalType := strings.TrimSpace(r.URL.Query().Get("signal_type"))
		if source == "" {
			source = "open_meteo"
		}
		locationKey := strings.TrimSpace(r.URL.Query().Get("location_key"))
		worldSignalLimit, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("world_signal_limit")))

		result, err := h.Usecase.Analyze(r.Context(), userID, deferenceapp.DistributionInput{
			Dataset:          deferenceapp.DistributionDataset(dataset),
			Axis:             axis,
			TagIDs:           tagIDs,
			Options:          options,
			Source:           source,
			LocationKey:      locationKey,
			SignalType:       signalType,
			WorldSignalLimit: worldSignalLimit,
		})
		if err != nil {
			switch {
			case deferenceapp.IsDistributionValidationError(err),
				errors.Is(err, externalsignalapp.ErrWorldSignalWindowInvalid),
				errors.Is(err, inferenceapp.ErrActionLogSortInvalid),
				errors.Is(err, inferenceapp.ErrActionLogOrderInvalid):
				sharedhttp.JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			default:
				sharedhttp.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to analyze distribution", err.Error())
			}
			return
		}

		sharedhttp.JSONSuccess(w, http.StatusOK, result, "distribution analysis fetched")
	}
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

func parseActionLogListOptions(r *http.Request) (inferenceapp.ActionLogListOptions, error) {
	var options inferenceapp.ActionLogListOptions
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
