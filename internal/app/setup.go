package app

import (
	"crypto/rsa"
	"database/sql"
	"livoir-blog/internal/delivery/http"
	"livoir-blog/internal/repository"
	"livoir-blog/internal/usecase"
	"livoir-blog/pkg/common"
	"livoir-blog/pkg/database"
	"livoir-blog/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

func SetupRouter(db *sql.DB, oauthConfig *oauth2.Config, privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey, encryptionKey string) (*gin.Engine, error) {
	if db == nil {
		logger.Log.Error("Database connection is nil")
		return nil, common.NewCustomError(500, "Database connection is nil")
	}
	if oauthConfig == nil {
		logger.Log.Error("OAuth config is nil")
		return nil, common.NewCustomError(500, "OAuth config is nil")
	}

	if encryptionKey == "" {
		logger.Log.Error("Encryption key is empty")
		return nil, common.NewCustomError(500, "Encryption key is required")
	}

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

	postUsecase, err := usecase.NewPostUsecase(postRepo, postVersionRepo, transactor)
	if err != nil {
		logger.Log.Error("Failed to initialize post usecase", zap.Error(err))
		return nil, err
	}
	categoryUsecase, err := usecase.NewCategoryUsecase(transactor, categoryRepo, postVersionRepo)
	if err != nil {
		logger.Log.Error("Failed to initialize category usecase", zap.Error(err))
		return nil, err
	}

	oauthUsecase, err := usecase.NewOauthUsecase(oauthRepo, tokenRepo, administratorRepo, administratorSessionRepo, transactor, encryptionKey)
	if err != nil {
		logger.Log.Error("Failed to initialize oauth usecase", zap.Error(err))
		return nil, err
	}

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	postsApi := r.Group("/posts")
	{
		http.NewPostHandler(postsApi, postUsecase)
	}
	categoriesApi := r.Group("/categories")
	{
		http.NewCategoryHandler(categoriesApi, categoryUsecase)
	}
	auth := r.Group("/auth")
	{
		http.NewAuthHandler(auth, oauthUsecase)
	}

	return r, nil
}
