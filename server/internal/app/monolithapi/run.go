package monolithapi

import (
	"log"
	"net/http"
	"os"
	"strings"

	identityapp "server/internal/modules/identity/app"
	identitydb "server/internal/modules/identity/adapters/db"
	identityhttp "server/internal/modules/identity/adapters/http"
	deferenceapp "server/internal/modules/deference/app"
	deferencedb "server/internal/modules/deference/adapters/db"
	deferencedistribution "server/internal/modules/deference/adapters/distribution_client"
	deferenceexternal "server/internal/modules/deference/adapters/external_signal_client"
	deferenceinference "server/internal/modules/deference/adapters/inference_client"
	deferencehttp "server/internal/modules/deference/adapters/http"
	externalsignalapp "server/internal/modules/externalsignal/app"
	externalsignaldb "server/internal/modules/externalsignal/adapters/db"
	externalsignalhttp "server/internal/modules/externalsignal/adapters/http"
	inferenceapp "server/internal/modules/inference/app"
	inferencedb "server/internal/modules/inference/adapters/db"
	inferencehttp "server/internal/modules/inference/adapters/http"
	shareddb "server/internal/shared/db"
	sharedhttp "server/internal/shared/http"
)

type deferenceHandlers struct {
	analysisSnapshotHandler     *deferencehttp.AnalysisSnapshotHandler
	analysisContextHandler      *deferencehttp.AnalysisContextHandler
	distributionAnalysisHandler *deferencehttp.DistributionAnalysisHandler
}

func Run() {
	conn, err := shareddb.NewPostgresConnection()
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer conn.Close()

	if err := shareddb.RunCoreMigrations(conn); err != nil {
		log.Fatalf("core migration failed: %v", err)
	}

	if isDeferenceMigrationsEnabled() {
		if err := shareddb.RunDeferenceMigrations(conn); err != nil {
			log.Fatalf("deference migration failed: %v", err)
		}
		log.Println("deference migrations enabled")
	} else {
		log.Println("deference migrations disabled")
	}

	if err := shareddb.RunTestDataMigrations(conn); err != nil {
		log.Fatalf("test-data migration failed: %v", err)
	}

	userRepo := identitydb.NewUserRepository(conn)
	authProviderRepo := identitydb.NewUserAuthProviderRepository(conn)
	refreshRepo := identitydb.NewRefreshTokenRepository(conn)
	userUsecase := &identityapp.UserService{
		Repo:             userRepo,
		AuthProviderRepo: authProviderRepo,
		RefreshRepo:      refreshRepo,
	}
	userHandler := &identityhttp.UserHandler{Usecase: userUsecase}

	targetRepo := inferencedb.NewTargetRepository(conn)
	targetUsecase := &inferenceapp.TargetService{Repo: targetRepo}
	targetHandler := &inferencehttp.TargetHandler{Usecase: targetUsecase}

	tagRepo := inferencedb.NewTagRepository(conn)
	tagUsecase := &inferenceapp.TagService{Repo: tagRepo}
	tagHandler := &inferencehttp.TagHandler{Usecase: tagUsecase}
	targetUsecase.TagRepo = tagRepo

	prototypeRepo := inferencedb.NewPrototypeRepository(conn)
	prototypeUsecase := &inferenceapp.PrototypeService{
		Repo:    prototypeRepo,
		TagRepo: tagRepo,
	}
	prototypeHandler := &inferencehttp.PrototypeHandler{Usecase: prototypeUsecase}

	actionLogRepo := inferencedb.NewActionLogRepository(conn)
	actionLogUsecase := &inferenceapp.ActionLogService{
		Repo:          actionLogRepo,
		TagRepo:       tagRepo,
		PrototypeRepo: prototypeRepo,
	}
	actionLogHandler := &inferencehttp.ActionLogHandler{Usecase: actionLogUsecase}

	analysisSnapshotRepo := deferencedb.NewAnalysisSnapshotRepository(conn)

	worldSignalRepo := externalsignaldb.NewWorldSignalRepository(conn)
	worldSignalUsecase := &externalsignalapp.QueryService{Repo: worldSignalRepo}
	worldSignalHandler := &externalsignalhttp.WorldSignalHandler{Usecase: worldSignalUsecase}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"message": "VAproject API Server", "oauth_login": "/api/v1/auth/oauth/google/login"}`))
	})
	mux.HandleFunc("/message", sharedhttp.Handler)

	mux.HandleFunc("/api/v1/auth/register", userHandler.RegisterHandler())
	mux.HandleFunc("/api/v1/auth/login", userHandler.LoginHandler())
	mux.HandleFunc("/api/v1/auth/me", userHandler.MeHandler())
	mux.HandleFunc("/api/v1/auth/refresh", userHandler.RefreshHandler())
	mux.HandleFunc("/api/v1/auth/profile", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			userHandler.GetProfileHandler()(w, r)
		} else if r.Method == http.MethodPut {
			userHandler.UpdateProfileHandler()(w, r)
		} else {
			sharedhttp.JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET and PUT only", nil)
		}
	})

	oauthHandler := &identityhttp.OAuthHandler{Usecase: userUsecase}
	mux.HandleFunc("/api/v1/auth/oauth/google/login", oauthHandler.GoogleLoginHandler())
	mux.HandleFunc("/api/v1/auth/oauth/google/callback", oauthHandler.GoogleCallbackHandler())

	mux.HandleFunc("/api/v1/targets", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			targetHandler.ListTargetsHandler()(w, r)
		} else if r.Method == http.MethodPost {
			targetHandler.CreateTargetHandler()(w, r)
		} else {
			sharedhttp.JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET and POST only", nil)
		}
	})
	mux.HandleFunc("/api/v1/targets/", func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/action_logs") {
			if r.Method == http.MethodGet {
				targetHandler.ListActionLogsByTargetHandler()(w, r)
			} else {
				sharedhttp.JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET only", nil)
			}
			return
		}

		targetHandler.TargetDetailHandler()(w, r)
	})

	mux.HandleFunc("/api/v1/action_logs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			actionLogHandler.ListHandler()(w, r)
		} else if r.Method == http.MethodPost {
			actionLogHandler.CreateHandler()(w, r)
		} else {
			sharedhttp.JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET and POST only", nil)
		}
	})
	mux.HandleFunc("/api/v1/action_logs/", actionLogHandler.DetailHandler())
	mux.HandleFunc("/api/v1/world_signals", worldSignalHandler.ListHandler())
	mux.HandleFunc("/api/v1/world_signals/fetch", worldSignalHandler.FetchOpenMeteoHandler())

	if isDeferenceEnabled() {
		registerDeferenceRoutes(mux, buildDeferenceHandlers(analysisSnapshotRepo, actionLogUsecase, worldSignalUsecase))
		log.Println("deference routes enabled")
	} else {
		log.Println("deference routes disabled")
	}

	mux.HandleFunc("/api/v1/tags", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			tagHandler.ListHandler()(w, r)
		} else if r.Method == http.MethodPost {
			tagHandler.CreateHandler()(w, r)
		} else {
			sharedhttp.JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET and POST only", nil)
		}
	})
	mux.HandleFunc("/api/v1/tags/", tagHandler.DetailHandler())

	mux.HandleFunc("/api/v1/prototypes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			prototypeHandler.ListHandler()(w, r)
		} else if r.Method == http.MethodPost {
			prototypeHandler.CreateHandler()(w, r)
		} else {
			sharedhttp.JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET and POST only", nil)
		}
	})
	mux.HandleFunc("/api/v1/prototypes/", prototypeHandler.DetailHandler())

	handlerWithAuth := jwtProtectedMux(mux)
	handlerWithCORS := corsMiddleware(handlerWithAuth)

	log.Println("server started on :8080")
	if err := http.ListenAndServe(":8080", handlerWithCORS); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "http://localhost:5173" || origin == "http://localhost:3000" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func jwtProtectedMux(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		publicPaths := []string{
			"/",
			"/message",
			"/api/v1/auth/register",
			"/api/v1/auth/login",
			"/api/v1/auth/refresh",
			"/api/v1/auth/oauth/google/login",
			"/api/v1/auth/oauth/google/callback",
		}
		for _, path := range publicPaths {
			if r.URL.Path == path {
				mux.ServeHTTP(w, r)
				return
			}
		}

		identityhttp.JWTMiddleware(mux).ServeHTTP(w, r)
	})
}

func isDeferenceEnabled() bool {
	return isFlagEnabled("DEFERENCE_ENABLED", true)
}

func isDeferenceMigrationsEnabled() bool {
	if _, ok := os.LookupEnv("DEFERENCE_MIGRATIONS_ENABLED"); ok {
		return isFlagEnabled("DEFERENCE_MIGRATIONS_ENABLED", true)
	}
	return isDeferenceEnabled()
}

func isFlagEnabled(envName string, defaultValue bool) bool {
	raw, ok := os.LookupEnv(envName)
	if !ok {
		return defaultValue
	}
	raw = strings.TrimSpace(strings.ToLower(raw))
	switch raw {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return defaultValue
	}
}

func buildDeferenceHandlers(
	analysisSnapshotRepo *deferencedb.AnalysisSnapshotRepository,
	actionLogUsecase *inferenceapp.ActionLogService,
	worldSignalUsecase *externalsignalapp.QueryService,
) *deferenceHandlers {
	analysisSnapshotUsecase := &deferenceapp.SnapshotService{Repo: analysisSnapshotRepo}
	analysisSnapshotHandler := &deferencehttp.AnalysisSnapshotHandler{Usecase: analysisSnapshotUsecase}

	inferenceQueryAdapter := &deferenceinference.MonolithInferenceQuery{Service: actionLogUsecase}
	externalSignalQueryAdapter := &deferenceexternal.MonolithExternalSignalQuery{Service: worldSignalUsecase}
	analysisContextUsecase := &deferenceapp.ContextService{
		InferenceQuery:      inferenceQueryAdapter,
		ExternalSignalQuery: externalSignalQueryAdapter,
	}
	analysisContextHandler := &deferencehttp.AnalysisContextHandler{Usecase: analysisContextUsecase}

	distributionAnalysisUsecase := &deferenceapp.DistributionService{
		Analyzer: deferencedistribution.NewMonolithDistributionAnalyzer(actionLogUsecase, worldSignalUsecase),
	}
	distributionAnalysisHandler := &deferencehttp.DistributionAnalysisHandler{Usecase: distributionAnalysisUsecase}

	return &deferenceHandlers{
		analysisSnapshotHandler:     analysisSnapshotHandler,
		analysisContextHandler:      analysisContextHandler,
		distributionAnalysisHandler: distributionAnalysisHandler,
	}
}

func registerDeferenceRoutes(mux *http.ServeMux, handlers *deferenceHandlers) {
	mux.HandleFunc("/api/v1/analysis_snapshots", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handlers.analysisSnapshotHandler.ListHandler()(w, r)
		} else if r.Method == http.MethodPost {
			handlers.analysisSnapshotHandler.CreateHandler()(w, r)
		} else if r.Method == http.MethodDelete {
			handlers.analysisSnapshotHandler.DeleteAllHandler()(w, r)
		} else {
			sharedhttp.JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "GET/POST/DELETE only", nil)
		}
	})
	mux.HandleFunc("/api/v1/analysis/context", handlers.analysisContextHandler.BuildHandler())
	mux.HandleFunc("/api/v1/analysis/distribution", handlers.distributionAnalysisHandler.AnalyzeHandler())
}
