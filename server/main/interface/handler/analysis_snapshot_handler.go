package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"server/main/domain"
	"server/main/usecases"
)

type AnalysisSnapshotHandler struct {
	Usecase *usecases.AnalysisSnapshotUsecase
}

type analysisSnapshotRequest struct {
	QueryText          string                    `json:"query_text"`
	Severity           string                    `json:"severity"`
	Score              int                       `json:"score"`
	DeltaAvgTag        float64                   `json:"delta_avg_tag"`
	DeltaVarTag        float64                   `json:"delta_var_tag"`
	DeltaPrototypeRate float64                   `json:"delta_prototype_rate"`
	PValue             *float64                  `json:"p_value"`
	Significant        *bool                     `json:"significant"`
	Current            domain.DistributionStats  `json:"current"`
	Baseline           *domain.DistributionStats `json:"baseline"`
}

func (h *AnalysisSnapshotHandler) ListHandler() http.HandlerFunc {
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
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit < 1 || limit > 100 {
			limit = 20
		}

		snapshots, err := h.Usecase.ListByUserID(r.Context(), userID, limit)
		if err != nil {
			JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch analysis snapshots", err.Error())
			return
		}
		JSONSuccess(w, http.StatusOK, map[string]interface{}{
			"snapshots": snapshots,
		}, "analysis snapshots fetched")
	}
}

func (h *AnalysisSnapshotHandler) CreateHandler() http.HandlerFunc {
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

		var req analysisSnapshotRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			JSONError(w, http.StatusBadRequest, "INVALID_JSON", "invalid json payload", err.Error())
			return
		}

		username, _ := r.Context().Value("username").(string)
		if username == "" {
			username = strconv.Itoa(userID)
		}

		snapshot := &domain.AnalysisSnapshot{
			UserID:             userID,
			QueryText:          req.QueryText,
			Severity:           req.Severity,
			Score:              req.Score,
			DeltaAvgTag:        req.DeltaAvgTag,
			DeltaVarTag:        req.DeltaVarTag,
			DeltaPrototypeRate: req.DeltaPrototypeRate,
			PValue:             req.PValue,
			Significant:        req.Significant,
			Current:            req.Current,
			Baseline:           req.Baseline,
			CreateUser:         username,
		}

		created, err := h.Usecase.Create(r.Context(), snapshot)
		if err != nil {
			switch {
			case errors.Is(err, usecases.ErrSnapshotSeverityInvalid),
				errors.Is(err, usecases.ErrSnapshotScoreInvalid),
				errors.Is(err, usecases.ErrSnapshotCurrentInvalid),
				errors.Is(err, usecases.ErrSnapshotCreateUserEmpty):
				JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
				return
			default:
				JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create analysis snapshot", err.Error())
				return
			}
		}

		JSONSuccess(w, http.StatusCreated, created, "analysis snapshot created")
	}
}

func (h *AnalysisSnapshotHandler) DeleteAllHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "DELETE method only", nil)
			return
		}
		userID, ok := currentUserID(r)
		if !ok {
			JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
			return
		}

		if err := h.Usecase.DeleteAllByUserID(r.Context(), userID); err != nil {
			JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to clear analysis snapshots", err.Error())
			return
		}
		JSONSuccess(w, http.StatusOK, nil, "analysis snapshots cleared")
	}
}
