package usecase

import (
	"context"
	"fmt"
	"livoir-blog/internal/domain"
)

type OAuthUsecase struct {
	oauthRepo domain.OAuthRepository
}

func NewOauthUsecase(oauthRepo domain.OAuthRepository) (domain.OAuthUsecase, error) {
	if oauthRepo == nil {
		return nil, fmt.Errorf("oauth repository is nil")
	}
	return &OAuthUsecase{
		oauthRepo: oauthRepo,
	}, nil
}

func (uc *OAuthUsecase) GetRedirectLoginUrl(ctx context.Context, state string) string {
	return uc.oauthRepo.GetRedirectLoginUrl(ctx, state)
}

func (uc *OAuthUsecase) LoginCallback(ctx context.Context, code string) (*domain.OAuthUser, error) {
	oauthUser, err := uc.oauthRepo.GetLoggedInUser(ctx, code)
	if err != nil {
		return nil, err
	}
	return oauthUser, err
}
