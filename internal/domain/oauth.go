package domain

import "context"

type OAuthUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

type OAuthUserResponse struct {
	User         *OAuthUser `json:"user"`
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
}

type OAuthRepository interface {
	GetRedirectLoginUrl(ctx context.Context, state string) string
	GetLoggedInUser(ctx context.Context, code string) (*OAuthUser, error)
}

type OAuthUsecase interface {
	GetRedirectLoginUrl(ctx context.Context, state string) string
	LoginCallback(ctx context.Context, code string) (*OAuthUserResponse, error)
}
