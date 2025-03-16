package app

import (
	"crypto/rsa"
	"database/sql"
	"livoir-blog/internal/domain"
	"livoir-blog/internal/repository"
	"livoir-blog/pkg/database"
	"livoir-blog/pkg/logger"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

type RepositoryProvider struct {
	Transactor                     domain.Transactor
	OAuthGoogleRepository          domain.OAuthRepository
	OAuthDiscordRepository         domain.OAuthRepository
	TokenRepository                domain.TokenRepository
	AdministratorRepository        domain.AdministratorRepository
	AdministratorSessionRepository domain.AdministratorSessionRepository
	PostRepository                 domain.PostRepository
	PostVersionRepository          domain.PostVersionRepository
	CategoryRepository             domain.CategoryRepository
}

func NewRepositoryProvider(db *sql.DB, oauthGoogleConfig, oauthDiscordConfig *oauth2.Config, privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) (*RepositoryProvider, error) {
	postRepo, err := repository.NewPostRepository(db)
	if err != nil {
		logger.Log.Error("Failed to initialize post repository", zap.Error(err))
		return nil, err
	}
	postVersionRepo, err := repository.NewPostVersionRepository(db)
	if err != nil {
		logger.Log.Error("Failed to initialize post version repository", zap.Error(err))
		return nil, err
	}
	oauthGoogleRepo, err := repository.NewOauthGoogleRepository(oauthGoogleConfig)
	if err != nil {
		logger.Log.Error("Failed to initialize oauth repository", zap.Error(err))
		return nil, err
	}
	oauthDiscordRepo, err := repository.NewOauthDiscordRepository(oauthDiscordConfig)
	if err != nil {
		logger.Log.Error("Failed to initialize oauth repository", zap.Error(err))
		return nil, err
	}
	transactor, err := database.NewSQLTransactor(db)
	if err != nil {
		logger.Log.Error("Failed to initialize transactor", zap.Error(err))
		return nil, err
	}
	categoryRepo, err := repository.NewCategoryRepository(db)
	if err != nil {
		logger.Log.Error("Failed to initialize category repository", zap.Error(err))
		return nil, err
	}
	tokenRepo, err := repository.NewTokenJWTRepository(privateKey, publicKey)
	if err != nil {
		logger.Log.Error("Failed to initialize token repository", zap.Error(err))
		return nil, err
	}
	administratorRepo, err := repository.NewAdministratorRepository(db)
	if err != nil {
		logger.Log.Error("Failed to initialize administrator repository", zap.Error(err))
		return nil, err
	}
	administratorSessionRepo, err := repository.NewAdministratorSessionRepository(db)
	if err != nil {
		logger.Log.Error("Failed to initialize administrator session repository", zap.Error(err))
		return nil, err
	}

	return &RepositoryProvider{
		transactor,
		oauthGoogleRepo,
		oauthDiscordRepo,
		tokenRepo,
		administratorRepo,
		administratorSessionRepo,
		postRepo,
		postVersionRepo,
		categoryRepo,
	}, nil
}

func (rp *RepositoryProvider) SetOauthRepositories(oauthGoogleRepo, oauthDiscordRepo domain.OAuthRepository) {
	rp.OAuthGoogleRepository = oauthGoogleRepo
	rp.OAuthDiscordRepository = oauthDiscordRepo
}
