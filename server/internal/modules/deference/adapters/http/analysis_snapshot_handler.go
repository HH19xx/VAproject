package httpadapter

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	deferenceapp "server/internal/modules/deference/app"
	deferencedomain "server/internal/modules/deference/domain"
	sharedhttp "server/internal/shared/http"
)

type AnalysisSnapshotHandler struct {
	Usecase *deferenceapp.SnapshotService
}

type analysisSnapshotRequest struct {
	QueryText          string                           `json:"query_text"`
	Severity           string                           `json:"severity"`
	Score              int                              `json:"score"`
	DeltaAvgTag        float64                          `json:"delta_avg_tag"`
	DeltaVarTag        float64                          `json:"delta_var_tag"`
	DeltaPrototypeRate float64                          `json:"delta_prototype_rate"`
	PValue             *float64                         `json:"p_value"`
	Significant        *bool                            `json:"significant"`
	Current            deferencedomain.DistributionStats `json:"current"`
	Baseline           *deferencedomain.DistributionStats `json:"baseline"`
}

func (h *AnalysisSnapshotHandler) ListHandler() http.HandlerFunc {
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
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit < 1 || limit > 100 {
			limit = 20
		}

		snapshots, err := h.Usecase.ListByUserID(r.Context(), userID, limit)
		if err != nil {
			sharedhttp.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch analysis snapshots", err.Error())
			return
		}
		sharedhttp.JSONSuccess(w, http.StatusOK, map[string]interface{}{
			"snapshots": snapshots,
		}, "analysis snapshots fetched")
	}
}

func (h *AnalysisSnapshotHandler) CreateHandler() http.HandlerFunc {
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

		var req analysisSnapshotRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sharedhttp.JSONError(w, http.StatusBadRequest, "INVALID_JSON", "invalid json payload", err.Error())
			return
		}

		username, _ := r.Context().Value("username").(string)
		if username == "" {
			username = strconv.Itoa(userID)
		}

		snapshot := &deferencedomain.AnalysisSnapshot{
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
			case errors.Is(err, deferenceapp.ErrSnapshotSeverityInvalid),
				errors.Is(err, deferenceapp.ErrSnapshotScoreInvalid),
				errors.Is(err, deferenceapp.ErrSnapshotCurrentInvalid),
				errors.Is(err, deferenceapp.ErrSnapshotCreateUserEmpty):
				sharedhttp.JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			default:
				sharedhttp.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create analysis snapshot", err.Error())
			}
			return
		}

		sharedhttp.JSONSuccess(w, http.StatusCreated, created, "analysis snapshot created")
	}
}

func (h *AnalysisSnapshotHandler) DeleteAllHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			sharedhttp.JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "DELETE method only", nil)
			return
		}
		userID, ok := currentUserID(r)
		if !ok {
			sharedhttp.JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication required", nil)
			return
		}

		if err := h.Usecase.DeleteAllByUserID(r.Context(), userID); err != nil {
			sharedhttp.JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to clear analysis snapshots", err.Error())
			return
		}
		sharedhttp.JSONSuccess(w, http.StatusOK, nil, "analysis snapshots cleared")
	}
}
