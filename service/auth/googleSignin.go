package auth

import (
	"github.com/kidusshun/planLog/config"
	"golang.org/x/oauth2/google"

	"golang.org/x/oauth2"
)

var GoogleOauthConfig = &oauth2.Config{
	ClientID:     config.GoogleEnvs.GoogleClientID,
	ClientSecret: config.GoogleEnvs.GoogleClientSecret,
	RedirectURL:  "http://localhost:8080/google/callback",
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
	Endpoint:     google.Endpoint,
}

const oauthGoogleURLAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="
