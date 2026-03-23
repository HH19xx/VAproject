package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"server/main/usecases"
)

type AnalysisContextHandler struct {
	Usecase *usecases.AnalysisContextUsecase
}

func (h *AnalysisContextHandler) BuildHandler() http.HandlerFunc {
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

		locationKey := strings.TrimSpace(r.URL.Query().Get("location_key"))
		source := strings.TrimSpace(r.URL.Query().Get("source"))
		signalType := strings.TrimSpace(r.URL.Query().Get("signal_type"))
		if source == "" {
			source = "open_meteo"
		}
		if locationKey == "" {
			JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", "location_key is required", nil)
			return
		}
		from, to, err := parseFromToRange(r)
		if err != nil {
			JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			return
		}
		limit, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("limit")))
		if limit < 1 || limit > 500 {
			limit = 200
		}

		result, err := h.Usecase.Build(r.Context(), userID, source, locationKey, signalType, from, to, limit)
		if err != nil {
			switch {
			case errors.Is(err, usecases.ErrWorldSignalWindowInvalid),
				errors.Is(err, usecases.ErrActionLogSortInvalid),
				errors.Is(err, usecases.ErrActionLogOrderInvalid):
				JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			default:
				JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to build analysis context", err.Error())
			}
			return
		}
		JSONSuccess(w, http.StatusOK, result, "analysis context fetched")
	}
}
