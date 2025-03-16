package auth

import (
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	GoogleOauthConfig *oauth2.Config
)

func NewGoogleOauthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     viper.GetString("auth.google.client_id"),
		ClientSecret: viper.GetString("auth.google.client_secret"),
		RedirectURL:  viper.GetString("auth.google.redirect_url"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}
