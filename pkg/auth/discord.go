package auth

import (
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

var (
	DiscordOauthConfig *oauth2.Config
)

func NewDiscordOauthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     viper.GetString("auth.discord.client_id"),
		ClientSecret: viper.GetString("auth.discord.client_secret"),
		RedirectURL:  viper.GetString("auth.discord.redirect_url"),
		Scopes:       []string{"identify", "email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discord.com/api/oauth2/authorize",
			TokenURL: "https://discord.com/api/oauth2/token",
		},
	}
}
