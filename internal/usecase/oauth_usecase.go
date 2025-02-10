package usecase

import (
	"context"
	"fmt"
	"livoir-blog/internal/domain"
	"livoir-blog/pkg/common"
	"time"
)

type OAuthUsecase struct {
	oauthRepo         domain.OAuthRepository
	tokenRepo         domain.TokenRepository
	administratorRepo domain.AdministratorRepository
}

func NewOauthUsecase(oauthRepo domain.OAuthRepository, tokenRepo domain.TokenRepository, administratorRepo domain.AdministratorRepository) (domain.OAuthUsecase, error) {
	if oauthRepo == nil {
		return nil, fmt.Errorf("oauth repository is nil")
	}
	if tokenRepo == nil {
		return nil, fmt.Errorf("token repository is nil")
	}
	if administratorRepo == nil {
		return nil, fmt.Errorf("administrator repository is nil")
	}

	return &OAuthUsecase{
		oauthRepo:         oauthRepo,
		tokenRepo:         tokenRepo,
		administratorRepo: administratorRepo,
	}, nil
}

func (uc *OAuthUsecase) GetRedirectLoginUrl(ctx context.Context, state string) string {
	return uc.oauthRepo.GetRedirectLoginUrl(ctx, state)
}

func (uc *OAuthUsecase) LoginCallback(ctx context.Context, code string) (*domain.OAuthUserResponse, error) {
	oauthUser, err := uc.oauthRepo.GetLoggedInUser(ctx, code)
	if err != nil {
		return nil, err
	}
	admin, err := uc.administratorRepo.FindByEmail(ctx, oauthUser.Email)
	if err != nil {
		return nil, err
	}
	if admin == nil {
		return nil, common.ErrUserNotFound
	}
	now := time.Now()
	expiredAt := now.Add(time.Minute * 10)
	accessToken, err := uc.tokenRepo.Generate(ctx, &domain.TokenData{
		UserID:    oauthUser.ID,
		Email:     oauthUser.Email,
		IssuedAt:  now.Unix(),
		ExpiredAt: expiredAt.Unix(),
	})
	if err != nil {
		return nil, err
	}
	expiredAt = now.Add(time.Hour * 24 * 7)
	refreshToken, err := uc.tokenRepo.Generate(ctx, &domain.TokenData{
		UserID:    oauthUser.ID,
		Email:     oauthUser.Email,
		IssuedAt:  now.Unix(),
		ExpiredAt: expiredAt.Unix(),
	})
	if err != nil {
		return nil, err
	}
	oauthUserResponse := &domain.OAuthUserResponse{
		User:         oauthUser,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	return oauthUserResponse, err
}
