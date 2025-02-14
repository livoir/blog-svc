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

type LoginCallbackRequest struct {
	Code      string `json:"code"`
	IpAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
}

type OAuthRepository interface {
	GetRedirectLoginUrl(ctx context.Context, state string) string
	GetLoggedInUser(ctx context.Context, code string) (*OAuthUser, error)
}

type OAuthUsecase interface {
	GetRedirectLoginUrl(ctx context.Context, state string) string
	LoginCallback(ctx context.Context, request *LoginCallbackRequest) (*OAuthUserResponse, error)
}
