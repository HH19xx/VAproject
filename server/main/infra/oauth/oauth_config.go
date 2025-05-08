package oauth

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var GoogleOAuthConfig = &oauth2.Config{
	ClientID:     "500930699796-doc4l3l8pknb61c006fksn0bnfg5i53h.apps.googleusercontent.com",
	ClientSecret: "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",
	RedirectURL:  "http://localhost:8080/auth/google/callback",
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
	Endpoint:     google.Endpoint,
}
