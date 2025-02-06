package repository

import (
	"context"
	"crypto/rsa"
	"livoir-blog/internal/domain"
	"livoir-blog/pkg/common"

	"github.com/golang-jwt/jwt/v5"
)

type TokenJWTRepository struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func NewTokenJWTRepository(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) (domain.TokenRepository, error) {
	if privateKey == nil {
		return nil, common.NewCustomError(500, "private key is nil")
	}
	if publicKey == nil {
		return nil, common.NewCustomError(500, "public key is nil")
	}
	return &TokenJWTRepository{
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

func (t *TokenJWTRepository) Generate(ctx context.Context, data *domain.TokenData) (string, error) {
	claims := jwt.MapClaims{
		"user_id": data.UserID,
		"email":   data.Email,
		"iat":     data.IssuedAt,
		"exp":     data.ExpiredAt,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(t.privateKey)
}

func (t *TokenJWTRepository) Validate(ctx context.Context, tokenStr string) (*domain.TokenData, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, common.ErrInvalidSigningMethod
		}
		return t.publicKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, common.ErrInvalidToken
	}
	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, common.ErrInvalidToken
	}
	email, ok := claims["email"].(string)
	if !ok {
		return nil, common.ErrInvalidToken
	}
	issuedAt, ok := claims["iat"].(float64)
	if !ok {
		return nil, common.ErrInvalidToken
	}
	expiredAt, ok := claims["exp"].(float64)
	if !ok {
		return nil, common.ErrInvalidToken
	}
	return &domain.TokenData{
		UserID:    userID,
		Email:     email,
		IssuedAt:  int64(issuedAt),
		ExpiredAt: int64(expiredAt),
	}, nil
}
