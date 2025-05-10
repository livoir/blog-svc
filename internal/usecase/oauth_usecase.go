package usecase

import (
	"context"
	"fmt"
	"livoir-blog/internal/domain"
	"livoir-blog/pkg/common"
	"livoir-blog/pkg/encryption"
	"livoir-blog/pkg/logger"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type OAuthUsecase struct {
	oauthRepo                  domain.OAuthRepository
	tokenRepo                  domain.TokenRepository
	administratorRepo          domain.AdministratorRepository
	administratorSessionRepo   domain.AdministratorSessionRepository
	cacheRepository            domain.CacheRepository
	txRepository               domain.Transactor
	encryptionKey              string
	accessTokenExpirationTime  time.Duration
	refreshTokenExpirationTime time.Duration
	tracer                     trace.Tracer
}

func NewOauthUsecase(oauthRepo domain.OAuthRepository,
	tokenRepo domain.TokenRepository,
	administratorRepo domain.AdministratorRepository,
	administratorSessionRepo domain.AdministratorSessionRepository,
	cacheRepository domain.CacheRepository,
	txRepository domain.Transactor, encryptionKey string,
	accessTokenExpirationTime, refreshTokenExpirationTime time.Duration) (domain.OAuthUsecase, error) {
	if oauthRepo == nil {
		return nil, fmt.Errorf("oauth repository is nil")
	}
	if tokenRepo == nil {
		return nil, fmt.Errorf("token repository is nil")
	}
	if administratorRepo == nil {
		return nil, fmt.Errorf("administrator repository is nil")
	}
	if administratorSessionRepo == nil {
		return nil, fmt.Errorf("administrator session repository is nil")
	}
	if txRepository == nil {
		return nil, fmt.Errorf("transaction repository is nil")
	}

	return &OAuthUsecase{
		oauthRepo:                  oauthRepo,
		tokenRepo:                  tokenRepo,
		administratorRepo:          administratorRepo,
		administratorSessionRepo:   administratorSessionRepo,
		txRepository:               txRepository,
		encryptionKey:              encryptionKey,
		accessTokenExpirationTime:  accessTokenExpirationTime,
		refreshTokenExpirationTime: refreshTokenExpirationTime,
		cacheRepository:            cacheRepository,
		tracer:                     otel.Tracer("oauth_usecase"),
	}, nil
}

func (uc *OAuthUsecase) GetRedirectLoginUrl(ctx context.Context, state string) string {
	return uc.oauthRepo.GetRedirectLoginUrl(ctx, state)
}

func (uc *OAuthUsecase) LoginCallback(ctx context.Context, request *domain.LoginCallbackRequest) (*domain.OAuthUserResponse, error) {
	tx, err := uc.txRepository.BeginTx()
	if err != nil {
		return nil, err
	}
	defer func(tx domain.Transaction) {
		if p := recover(); p != nil {
			e := tx.Rollback()
			if e != nil {
				logger.Log.Error("Failed to rollback transaction", zap.Error(e), zap.String("error_source", "panic_recovery"))
			}
			panic(p)
		} else if err != nil {
			e := tx.Rollback()
			if e != nil {
				logger.Log.Error("Failed to rollback transaction", zap.Error(e), zap.String("error_source", "error_propagation"))
			}
		}
	}(tx)
	oauthUser, err := uc.oauthRepo.GetLoggedInUser(ctx, request.Code)
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
	expiredAt := now.Add(uc.accessTokenExpirationTime)
	accessToken, err := uc.tokenRepo.Generate(ctx, &domain.TokenData{
		UserID:    oauthUser.ID,
		Email:     oauthUser.Email,
		IssuedAt:  now.Unix(),
		ExpiredAt: expiredAt.Unix(),
	})
	if err != nil {
		return nil, err
	}
	expiredAt = now.Add(uc.refreshTokenExpirationTime)
	refreshToken, err := uc.tokenRepo.Generate(ctx, &domain.TokenData{
		UserID:    oauthUser.ID,
		Email:     oauthUser.Email,
		IssuedAt:  now.Unix(),
		ExpiredAt: expiredAt.Unix(),
	})
	if err != nil {
		return nil, err
	}

	encryptedAccessToken, err := encryption.Encrypt(accessToken, []byte(uc.encryptionKey))
	encryptedRefreshToken, err := encryption.Encrypt(refreshToken, []byte(uc.encryptionKey))
	if err != nil {
		logger.Log.Error("Failed to encrypt refresh token", zap.Error(err))
		return nil, err
	}

	err = uc.administratorSessionRepo.Insert(ctx, tx, &domain.AdministratorSession{
		AdministratorID: admin.ID,
		EncryptedToken:  encryptedRefreshToken,
		IpAddress:       request.IpAddress,
		UserAgent:       request.UserAgent,
	})
	if err != nil {
		return nil, err
	}
	err = uc.cacheRepository.Set(ctx, fmt.Sprintf("oauth:%s:%s", oauthUser.ID, encryptedAccessToken), 1, uc.accessTokenExpirationTime)
	if err != nil {
		logger.Log.Error("Failed to set cache", zap.Error(err))
		return nil, err
	}
	oauthUserResponse := &domain.OAuthUserResponse{
		User:         oauthUser,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return oauthUserResponse, err
}
