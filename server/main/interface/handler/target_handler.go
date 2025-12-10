package handler

import (
	"encoding/json"
	"net/http"
	"server/main/app/usecases"
	"strconv"
	"strings"
)

// 観察対象に関するHTTPハンドラを提供
type TargetHandler struct {
	Usecase *usecases.TargetUsecase
}

// 観察対象の一覧を取得（GET /api/v1/targets）
func (h *TargetHandler) ListTargetsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// GETメソッドのみ許可
		if r.Method != http.MethodGet {
			JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "このエンドポイントはGETメソッドのみ対応しています", nil)
			return
		}

		// クエリパラメータからページとリミットを取得（デフォルト: page=1, limit=20）
		pageStr := r.URL.Query().Get("page")
		limitStr := r.URL.Query().Get("limit")

		page, _ := strconv.Atoi(pageStr)
		if page < 1 {
			page = 1
		}

		limit, _ := strconv.Atoi(limitStr)
		if limit < 1 || limit > 100 {
			limit = 20
		}

		// ユースケースから観察対象一覧を取得
		targets, total, err := h.Usecase.GetTargets(r.Context(), page, limit)
		if err != nil {
			JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "観察対象の取得に失敗しました", err.Error())
			return
		}

		// レスポンスボディを構築
		response := map[string]interface{}{
			"targets": targets,
			"pagination": map[string]interface{}{
				"page":  page,
				"limit": limit,
				"total": total,
			},
		}

		JSONSuccess(w, http.StatusOK, response, "観察対象の一覧を取得しました")
	}
}

// 観察対象を新規作成（POST /api/v1/targets）
func (h *TargetHandler) CreateTargetHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// POSTメソッドのみ許可
		if r.Method != http.MethodPost {
			JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "このエンドポイントはPOSTメソッドのみ対応しています", nil)
			return
		}

		// リクエストボディをパース
		var req struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			JSONError(w, http.StatusBadRequest, "INVALID_JSON", "JSONの形式が正しくありません", err.Error())
			return
		}

		// JWTミドルウェアから取得したユーザー名を使用
		username, ok := r.Context().Value("username").(string)
		if !ok || username == "" {
			username = "system" // フォールバック
		}

		// ユースケースで観察対象を作成
		target, err := h.Usecase.CreateTarget(r.Context(), req.Name, req.Description, username)
		if err != nil {
			// バリデーションエラー
			if err == usecases.ErrTargetNameRequired || err == usecases.ErrTargetNameTooLong {
				JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
				return
			}
			// その他のエラー（DB制約違反など）
			JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "観察対象の作成に失敗しました", err.Error())
			return
		}

		JSONSuccess(w, http.StatusCreated, target, "観察対象を作成しました")
	}
}

// 個別の観察対象に対する操作を処理（GET/PUT/DELETE /api/v1/targets/{id}）
func (h *TargetHandler) TargetDetailHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// URLパスからIDを抽出
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/targets/")
		idStr := strings.Split(path, "/")[0]

		id, err := strconv.Atoi(idStr)
		if err != nil || id < 1 {
			JSONError(w, http.StatusBadRequest, "INVALID_ID", "IDの形式が正しくありません", nil)
			return
		}

		// HTTPメソッドに応じて処理を分岐
		switch r.Method {
		case http.MethodGet:
			h.getTargetByID(w, r, id)
		case http.MethodPut:
			h.updateTarget(w, r, id)
		case http.MethodDelete:
			h.deleteTarget(w, r, id)
		default:
			JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "このエンドポイントはGET、PUT、DELETEメソッドのみ対応しています", nil)
		}
	}
}

// 指定IDの観察対象を取得する内部メソッド
func (h *TargetHandler) getTargetByID(w http.ResponseWriter, r *http.Request, id int) {
	target, err := h.Usecase.GetTargetByID(r.Context(), id)
	if err != nil {
		JSONError(w, http.StatusNotFound, "NOT_FOUND", "指定された観察対象が見つかりません", nil)
		return
	}

	JSONSuccess(w, http.StatusOK, target, "観察対象を取得しました")
}

// 指定IDの観察対象を更新する内部メソッド
func (h *TargetHandler) updateTarget(w http.ResponseWriter, r *http.Request, id int) {
	// リクエストボディをパース
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_JSON", "JSONの形式が正しくありません", err.Error())
		return
	}

	// JWTミドルウェアから取得したユーザー名を使用
	username, ok := r.Context().Value("username").(string)
	if !ok || username == "" {
		username = "system"
	}

	// ユースケースで観察対象を更新
	target, err := h.Usecase.UpdateTarget(r.Context(), id, req.Name, req.Description, username)
	if err != nil {
		// バリデーションエラー
		if err == usecases.ErrTargetNameRequired || err == usecases.ErrTargetNameTooLong {
			JSONError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
			return
		}
		// 対象が見つからない
		JSONError(w, http.StatusNotFound, "NOT_FOUND", "指定された観察対象が見つかりません", nil)
		return
	}

	JSONSuccess(w, http.StatusOK, target, "観察対象を更新しました")
}

// 指定IDの観察対象を論理削除する内部メソッド
func (h *TargetHandler) deleteTarget(w http.ResponseWriter, r *http.Request, id int) {
	// ユースケースで観察対象を論理削除
	err := h.Usecase.DeleteTarget(r.Context(), id)
	if err != nil {
		JSONError(w, http.StatusNotFound, "NOT_FOUND", "指定された観察対象が見つかりません", nil)
		return
	}

	JSONSuccess(w, http.StatusOK, nil, "観察対象を削除しました")
}
