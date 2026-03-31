package httpadapter

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	deferenceapp "server/internal/modules/deference/app"
	externalsignalapp "server/internal/modules/externalsignal/app"
	inferenceapp "server/internal/modules/inference/app"
	sharedhttp "server/internal/shared/http"
)

type AnalysisContextHandler struct {
	Usecase *deferenceapp.ContextService
}

func (h *AnalysisContextHandler) BuildHandler() http.HandlerFunc {
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

		locationKey := strings.TrimSpace(r.URL.Query().Get("location_key"))
		source := strings.TrimSpace(r.URL.Query().Get("source"))
		signalType := strings.TrimSpace(r.URL.Query().Get("signal_type"))
		if source == "" {
			source = "open_meteo"
		}
		if locationKey == "" {
			sharedhttp.JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", "location_key is required", nil)
			return
		}
		from, to, err := parseFromToRange(r)
		if err != nil {
			sharedhttp.JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			return
		}
		limit, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("limit")))
		if limit < 1 || limit > 500 {
			limit = 200
		}

		result, err := h.Usecase.Build(r.Context(), userID, source, locationKey, signalType, from, to, limit)
		if err != nil {
			switch {
			case errors.Is(err, externalsignalapp.ErrWorldSignalWindowInvalid),
				errors.Is(err, inferenceapp.ErrActionLogSortInvalid),
				errors.Is(err, inferenceapp.ErrActionLogOrderInvalid):
				sharedhttp.JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			default:
				sharedhttp.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to build analysis context", err.Error())
			}
			return
		}
		sharedhttp.JSONSuccess(w, http.StatusOK, result, "analysis context fetched")
	}
}

func parseFromToRange(r *http.Request) (time.Time, time.Time, error) {
	fromRaw := strings.TrimSpace(r.URL.Query().Get("from"))
	toRaw := strings.TrimSpace(r.URL.Query().Get("to"))

	if fromRaw == "" || toRaw == "" {
		now := time.Now().UTC()
		from := now.AddDate(0, 0, -7)
		return from, now, nil
	}

	from, err := time.Parse(time.RFC3339, fromRaw)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("from must be RFC3339")
	}
	to, err := time.Parse(time.RFC3339, toRaw)
	if err != nil {
		return time.Time{}, time.Time{}, errors.New("to must be RFC3339")
	}
	if from.After(to) {
		return time.Time{}, time.Time{}, externalsignalapp.ErrWorldSignalWindowInvalid
	}
	return from, to, nil
}
