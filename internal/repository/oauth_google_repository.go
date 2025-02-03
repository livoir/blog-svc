package repository

import (
	"context"
	"encoding/json"
	"livoir-blog/internal/domain"
	"livoir-blog/pkg/common"
	"livoir-blog/pkg/logger"
	"net/http"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

type OAuthGoogleRepository struct {
	GoogleOauthConfig *oauth2.Config
}

func NewOauthGoogleRepository(googleOauthConfig *oauth2.Config) (domain.OAuthRepository, error) {
	if googleOauthConfig == nil {
		logger.Log.Error("Google oauth config is nil")
		return nil, common.NewCustomError(http.StatusInternalServerError, "Google oauth config is nil")
	}

	return &OAuthGoogleRepository{
		GoogleOauthConfig: googleOauthConfig,
	}, nil
}

func (o *OAuthGoogleRepository) GetLoggedInUser(ctx context.Context, code string) (*domain.OAuthUser, error) {
	token, err := o.GoogleOauthConfig.Exchange(ctx, code)
	if err != nil {
		logger.Log.Error("Failed to exchange code for token", zap.Error(err))
		return nil, err
	}
	user, err := o.GoogleOauthConfig.Client(ctx, token).Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	var oauthUser domain.OAuthUser
	defer user.Body.Close()
	if err := json.NewDecoder(user.Body).Decode(&oauthUser); err != nil {
		return nil, err
	}
	return &oauthUser, nil
}

func (o *OAuthGoogleRepository) GetRedirectLoginUrl(ctx context.Context, state string) string {
	url := o.GoogleOauthConfig.AuthCodeURL(state, oauth2.ApprovalForce)
	return url
}
