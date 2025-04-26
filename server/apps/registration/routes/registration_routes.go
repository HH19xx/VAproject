package routes

import (
	"server/apps/registration/handler"

	"github.com/gorilla/mux"
)

// SetupRegistrationRoutes は登録関連のルートを設定します。
func SetupRegistrationRoutes(router *mux.Router, registrationHandler *handler.RegistrationHandler) {
	router.HandleFunc("/register", registrationHandler.RegisterUserHandler).Methods("POST")
	// 他に必要なルートがあればここに追加します。
}
