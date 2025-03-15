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

type OAuthDiscordRepository struct {
	DiscordOauthConfig *oauth2.Config
}

func NewOauthDiscordRepository(discordOauthConfig *oauth2.Config) (domain.OAuthRepository, error) {
	if discordOauthConfig == nil {
		logger.Log.Error("Discord oauth config is nil")
		return nil, common.NewCustomError(http.StatusInternalServerError, "Discord oauth config is required")
	}
	return &OAuthDiscordRepository{
		DiscordOauthConfig: discordOauthConfig,
	}, nil
}

func (o *OAuthDiscordRepository) GetLoggedInUser(ctx context.Context, code string) (*domain.OAuthUser, error) {
	token, err := o.DiscordOauthConfig.Exchange(ctx, code)
	if err != nil {
		logger.Log.Error("Failed to exchange code for token", zap.Error(err))
		return nil, err
	}
	user, err := o.DiscordOauthConfig.Client(ctx, token).Get("https://discord.com/api/users/@me")
	if err != nil {
		logger.Log.Error("Failed to get user info", zap.Error(err))
		return nil, err
	}
	var oauthUser domain.OAuthUser
	defer user.Body.Close()
	if err := json.NewDecoder(user.Body).Decode(&oauthUser); err != nil {
		logger.Log.Error("Failed to decode user info", zap.Error(err))
		return nil, err
	}
	return &oauthUser, nil
}

func (o *OAuthDiscordRepository) GetRedirectLoginUrl(ctx context.Context, state string) string {
	return o.DiscordOauthConfig.AuthCodeURL(state, oauth2.ApprovalForce)
}
