package auth

import (
	"golang.org/x/oauth2/google"

	"golang.org/x/oauth2"
)

var GoogleOauthConfig = &oauth2.Config{
	ClientID:     "769650932076-jsb8tnojaij51gredqukd7l2u3h55nrg.apps.googleusercontent.com",
	ClientSecret: "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",
	RedirectURL: "http://localhost:3000/api/auth/callback/google",
	Scopes:       []string{"https://www.googleapis.com/auth/calendar","https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
	Endpoint:     google.Endpoint,
}

// const oauthGoogleURLAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="
