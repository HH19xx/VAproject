package httpadapter

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	identityapp "server/internal/modules/identity/app"
	sharedhttp "server/internal/shared/http"
)

type UserHandler struct {
	Usecase *identityapp.UserService
}

func (h *UserHandler) RegisterHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name     string `json:"name"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("RegisterHandler decode error: %v", err)
			sharedhttp.JSONError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", nil)
			return
		}

		ctx := r.Context()
		userID, err := h.Usecase.Register(ctx, req.Name, req.Password)
		if err != nil {
			sharedhttp.JSONError(w, http.StatusBadRequest, "REGISTRATION_ERROR", err.Error(), nil)
			return
		}

		tokens, err := h.Usecase.IssueTokens(ctx, userID, "auth")
		if err != nil {
			sharedhttp.JSONError(w, http.StatusInternalServerError, "TOKEN_ERROR", "failed to issue tokens", nil)
			return
		}

		sharedhttp.JSONSuccess(w, http.StatusCreated, map[string]interface{}{
			"user_id":                 userID,
			"access_token":            tokens.AccessToken,
			"refresh_token":           tokens.RefreshToken,
			"access_token_expires_at": tokens.AccessTokenExpiresAt.UTC().Format(time.RFC3339),
		}, "user registered")
	}
}

func (h *UserHandler) LoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name     string `json:"name"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sharedhttp.JSONError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", nil)
			return
		}

		ctx := r.Context()
		userID, err := h.Usecase.Login(ctx, req.Name, req.Password)
		if err != nil {
			sharedhttp.JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "authentication failed", nil)
			return
		}

		tokens, err := h.Usecase.IssueTokens(ctx, userID, "auth")
		if err != nil {
			sharedhttp.JSONError(w, http.StatusInternalServerError, "TOKEN_ERROR", "failed to issue tokens", nil)
			return
		}

		sharedhttp.JSONSuccess(w, http.StatusOK, map[string]interface{}{
			"access_token":            tokens.AccessToken,
			"refresh_token":           tokens.RefreshToken,
			"access_token_expires_at": tokens.AccessTokenExpiresAt.UTC().Format(time.RFC3339),
		}, "login succeeded")
	}
}

func (h *UserHandler) MeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("userID").(int)
		sharedhttp.JSONSuccess(w, http.StatusOK, map[string]interface{}{
			"id": userID,
		}, "current user")
	}
}

func (h *UserHandler) GetProfileHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("userID").(int)
		ctx := r.Context()

		profile, err := h.Usecase.GetProfile(ctx, userID)
		if err != nil {
			sharedhttp.JSONError(w, http.StatusNotFound, "NOT_FOUND", "user not found", nil)
			return
		}

		sharedhttp.JSONSuccess(w, http.StatusOK, map[string]interface{}{
			"id":             profile.User.ID,
			"name":           profile.User.Name,
			"email":          profile.User.Email,
			"auth_providers": profile.AuthProviders,
		}, "profile loaded")
	}
}

func (h *UserHandler) UpdateProfileHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("userID").(int)

		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sharedhttp.JSONError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid request body", nil)
			return
		}

		ctx := r.Context()
		if err := h.Usecase.UpdateProfile(ctx, userID, req.Name); err != nil {
			sharedhttp.JSONError(w, http.StatusBadRequest, "UPDATE_ERROR", err.Error(), nil)
			return
		}

		sharedhttp.JSONSuccess(w, http.StatusOK, map[string]interface{}{
			"user_id": userID,
		}, "profile updated")
	}
}

func (h *UserHandler) RefreshHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RefreshToken == "" {
			sharedhttp.JSONError(w, http.StatusBadRequest, "BAD_REQUEST", "refresh_token is required", nil)
			return
		}

		tokens, err := h.Usecase.RefreshTokens(r.Context(), req.RefreshToken)
		if err != nil {
			sharedhttp.JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "refresh token is invalid", nil)
			return
		}

		sharedhttp.JSONSuccess(w, http.StatusOK, map[string]interface{}{
			"access_token":            tokens.AccessToken,
			"refresh_token":           tokens.RefreshToken,
			"access_token_expires_at": tokens.AccessTokenExpiresAt.UTC().Format(time.RFC3339),
		}, "token refreshed")
	}
}
