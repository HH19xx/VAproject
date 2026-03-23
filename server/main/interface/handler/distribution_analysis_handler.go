package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"server/main/usecases"
)

type DistributionAnalysisHandler struct {
	Usecase *usecases.DistributionAnalysisUsecase
}

func (h *DistributionAnalysisHandler) AnalyzeHandler() http.HandlerFunc {
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

		tagIDs, err := parseTagIDsCSV(r.URL.Query().Get("tag_ids"))
		if err != nil {
			JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", "tag_ids must be comma separated positive integers", nil)
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

		options, err := parseActionLogListOptions(r)
		if err != nil {
			JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			return
		}
		options.AnyTagIDs = anyTagIDs
		options.AnyTagGroups = anyTagGroups
		options.ExcludeTagIDs = excludeTagIDs

		dataset := strings.TrimSpace(r.URL.Query().Get("dataset"))
		if dataset == "" {
			dataset = string(usecases.DistributionDatasetActionLogs)
		}
		axis := strings.TrimSpace(r.URL.Query().Get("axis"))
		source := strings.TrimSpace(r.URL.Query().Get("source"))
		signalType := strings.TrimSpace(r.URL.Query().Get("signal_type"))
		if source == "" {
			source = "open_meteo"
		}
		locationKey := strings.TrimSpace(r.URL.Query().Get("location_key"))
		worldSignalLimit, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("world_signal_limit")))

		result, err := h.Usecase.Analyze(r.Context(), userID, usecases.DistributionAnalysisInput{
			Dataset:          usecases.DistributionDataset(dataset),
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
			case errors.Is(err, usecases.ErrDistributionDatasetInvalid),
				errors.Is(err, usecases.ErrDistributionAxisRequired),
				errors.Is(err, usecases.ErrDistributionAxisInvalid),
				errors.Is(err, usecases.ErrDistributionLocationNeeded),
				errors.Is(err, usecases.ErrDistributionWindowInvalid),
				errors.Is(err, usecases.ErrActionLogSortInvalid),
				errors.Is(err, usecases.ErrActionLogOrderInvalid),
				errors.Is(err, usecases.ErrWorldSignalWindowInvalid):
				JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			default:
				JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to analyze distribution", err.Error())
			}
			return
		}

		JSONSuccess(w, http.StatusOK, result, "distribution analysis fetched")
	}
}
