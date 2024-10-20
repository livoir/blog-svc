package app

import (
	"database/sql"
	"livoir-blog/internal/delivery/http"
	"livoir-blog/internal/repository"
	"livoir-blog/internal/usecase"

	"github.com/gin-gonic/gin"
)

func SetupRouter(db *sql.DB) *gin.Engine {
	postRepo := repository.NewPostRepository(db)
	postVersionRepo := repository.NewPostVersionRepository(db)
	postUsecase := usecase.NewPostUsecase(postRepo, postVersionRepo)

	r := gin.Default()
	http.NewPostHandler(r, postUsecase)

	return r
}
