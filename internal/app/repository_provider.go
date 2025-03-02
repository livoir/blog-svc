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
	transactor                     domain.Transactor
	oauthRepository                domain.OAuthRepository
	tokenRepository                domain.TokenRepository
	administratorRepository        domain.AdministratorRepository
	administratorSessionRepository domain.AdministratorSessionRepository
	postRepository                 domain.PostRepository
	postVersionRepository          domain.PostVersionRepository
	categoryRepository             domain.CategoryRepository
}

func NewRepositoryProvider(db *sql.DB, oauthConfig *oauth2.Config, privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey) (*RepositoryProvider, error) {
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
	oauthRepo, err := repository.NewOauthGoogleRepository(oauthConfig)
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
		oauthRepo,
		tokenRepo,
		administratorRepo,
		administratorSessionRepo,
		postRepo,
		postVersionRepo,
		categoryRepo,
	}, nil
}

func (rp *RepositoryProvider) SetOauthRepository(oauthRepo domain.OAuthRepository) {
	rp.oauthRepository = oauthRepo
}
