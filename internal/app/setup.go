package app

import (
	"database/sql"
	"livoir-blog/internal/delivery/http"
	"livoir-blog/internal/usecase"
	"livoir-blog/pkg/common"
	"livoir-blog/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func SetupRouter(db *sql.DB, repoProvider *RepositoryProvider, encryptionKey string, accessTokenExpiration time.Duration, refreshTokenExpiration time.Duration) (*gin.Engine, error) {
	if db == nil {
		logger.Log.Error("Database connection is nil")
		return nil, common.NewCustomError(500, "Database connection is nil")
	}
	if repoProvider == nil {
		logger.Log.Error("Repository provider is nil")
		return nil, common.NewCustomError(500, "Repository provider is required")
	}

	if encryptionKey == "" {
		logger.Log.Error("Encryption key is empty")
		return nil, common.NewCustomError(500, "Encryption key is required")
	}

	postUsecase, err := usecase.NewPostUsecase(repoProvider.postRepository, repoProvider.postVersionRepository, repoProvider.transactor)
	if err != nil {
		logger.Log.Error("Failed to initialize post usecase", zap.Error(err))
		return nil, err
	}
	categoryUsecase, err := usecase.NewCategoryUsecase(repoProvider.transactor, repoProvider.categoryRepository, repoProvider.postVersionRepository)
	if err != nil {
		logger.Log.Error("Failed to initialize category usecase", zap.Error(err))
		return nil, err
	}

	oauthUsecase, err := usecase.NewOauthUsecase(repoProvider.oauthRepository, repoProvider.tokenRepository, repoProvider.administratorRepository, repoProvider.administratorSessionRepository, repoProvider.transactor, encryptionKey, accessTokenExpiration, refreshTokenExpiration)
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
		http.NewAuthHandler(auth, oauthUsecase, accessTokenExpiration, refreshTokenExpiration)
	}

	return r, nil
}
