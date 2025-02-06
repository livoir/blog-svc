package domain

import "context"

type TokenData struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	IssuedAt  int64  `json:"iat"`
	ExpiredAt int64  `json:"exp"`
}

type GenerateTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type TokenRepository interface {
	Generate(ctx context.Context, data *TokenData) (string, error)
	Validate(ctx context.Context, tokenStr string) (*TokenData, error)
}
