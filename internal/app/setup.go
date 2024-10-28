package app

import (
	"database/sql"
	"livoir-blog/internal/delivery/http"
	"livoir-blog/internal/repository"
	"livoir-blog/internal/usecase"
	"livoir-blog/pkg/database"
	"livoir-blog/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func SetupRouter(db *sql.DB) (*gin.Engine, error) {
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
	transactor, err := database.NewSQLTransactor(db)
	if err != nil {
		logger.Log.Error("Failed to initialize transactor", zap.Error(err))
		return nil, err
	}
	postUsecase, err := usecase.NewPostUsecase(postRepo, postVersionRepo, transactor)
	if err != nil {
		logger.Log.Error("Failed to initialize post usecase", zap.Error(err))
		return nil, err
	}
	categoryRepo, err := repository.NewCategoryRepository(db)
	if err != nil {
		logger.Log.Error("Failed to initialize category repository", zap.Error(err))
		return nil, err
	}
	categoryUsecase, err := usecase.NewCategoryUsecase(transactor, categoryRepo, postVersionRepo)
	if err != nil {
		logger.Log.Error("Failed to initialize category usecase", zap.Error(err))
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

	return r, nil
}
