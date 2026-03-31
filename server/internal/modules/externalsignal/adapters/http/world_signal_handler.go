package httpadapter

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	externalsignalapp "server/internal/modules/externalsignal/app"
	sharedhttp "server/internal/shared/http"
)

type WorldSignalHandler struct {
	Usecase *externalsignalapp.QueryService
}

type fetchOpenMeteoRequest struct {
	Source       string  `json:"source"`
	LocationKey  string  `json:"location_key"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	PastDays     int     `json:"past_days"`
	ForecastDays int     `json:"forecast_days"`
}

func (h *WorldSignalHandler) ListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			sharedhttp.JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET method only", nil)
			return
		}

		locationKey := strings.TrimSpace(r.URL.Query().Get("location_key"))
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
		if limit < 1 || limit > 1000 {
			limit = 200
		}
		source := strings.TrimSpace(r.URL.Query().Get("source"))
		signalType := strings.TrimSpace(r.URL.Query().Get("signal_type"))
		if source == "" {
			source = "open_meteo"
		}

		signals, err := h.Usecase.ListByRange(r.Context(), source, locationKey, signalType, from, to, limit)
		if err != nil {
			switch {
			case errors.Is(err, externalsignalapp.ErrWorldSignalWindowInvalid):
				sharedhttp.JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			default:
				sharedhttp.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch world signals", err.Error())
			}
			return
		}

		sharedhttp.JSONSuccess(w, http.StatusOK, map[string]interface{}{
			"signals":      signals,
			"source":       source,
			"location_key": locationKey,
			"from":         from,
			"to":           to,
		}, "world signals fetched")
	}
}

func (h *WorldSignalHandler) FetchOpenMeteoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			sharedhttp.JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "POST method only", nil)
			return
		}
		if _, ok := currentUserID(r); !ok {
			sharedhttp.JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
			return
		}

		var req fetchOpenMeteoRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sharedhttp.JSONError(w, http.StatusBadRequest, "INVALID_JSON", "invalid json payload", err.Error())
			return
		}
		if req.PastDays == 0 {
			req.PastDays = 7
		}
		if req.ForecastDays == 0 {
			req.ForecastDays = 1
		}

		username, _ := r.Context().Value("username").(string)
		source := strings.TrimSpace(req.Source)
		if source == "" {
			source = "open_meteo"
		}
		result, err := h.Usecase.FetchAndStore(
			r.Context(),
			source,
			req.LocationKey,
			req.Latitude,
			req.Longitude,
			req.PastDays,
			req.ForecastDays,
			username,
		)
		if err != nil {
			switch {
			case errors.Is(err, externalsignalapp.ErrWorldSignalLocationRequired),
				errors.Is(err, externalsignalapp.ErrWorldSignalSourceUnsupported),
				errors.Is(err, externalsignalapp.ErrWorldSignalLatitudeInvalid),
				errors.Is(err, externalsignalapp.ErrWorldSignalLongitudeInvalid),
				errors.Is(err, externalsignalapp.ErrWorldSignalDaysInvalid):
				sharedhttp.JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			default:
				sharedhttp.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch world signal from open data", err.Error())
			}
			return
		}

		sharedhttp.JSONSuccess(w, http.StatusCreated, result, "world signals fetched and stored")
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
