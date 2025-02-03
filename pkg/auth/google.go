package auth

import (
	"context"
	"encoding/json"
	"livoir-blog/pkg/common"
	"net/http"

	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	GoogleOauthConfig *oauth2.Config
)

type GoogleUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

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

func GetUserInfo(token *oauth2.Token) (*GoogleUser, error) {
	client := GoogleOauthConfig.Client(context.Background(), token)
	response, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, common.NewCustomError(response.StatusCode, "Failed to get user info")
	}
	user := &GoogleUser{}
	if err := json.NewDecoder(response.Body).Decode(user); err != nil {
		return nil, err
	}

	return user, nil
}
