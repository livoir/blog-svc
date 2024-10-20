package app

import (
	"database/sql"
	"livoir-blog/internal/delivery/http"
	"livoir-blog/internal/repository"
	"livoir-blog/internal/usecase"
	"livoir-blog/pkg/database"

	"github.com/gin-gonic/gin"
)

func SetupRouter(db *sql.DB) *gin.Engine {
	postRepo := repository.NewPostRepository(db)
	postVersionRepo := repository.NewPostVersionRepository(db)
	transactor := database.NewSQLTransactor(db)
	postUsecase := usecase.NewPostUsecase(postRepo, postVersionRepo, transactor)

	r := gin.Default()
	http.NewPostHandler(r, postUsecase)

	return r
}
