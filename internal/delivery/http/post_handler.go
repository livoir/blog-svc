package http

import (
	"errors"
	"net/http"

	"livoir-blog/internal/domain"

	"github.com/gin-gonic/gin"
)

type PostHandler struct {
	PostUsecase domain.PostUsecase
}

func NewPostHandler(r *gin.Engine, usecase domain.PostUsecase) {
	handler := &PostHandler{
		PostUsecase: usecase,
	}
	posts := r.Group("/posts")
	{
		posts.GET("/:id", handler.GetPost)
		posts.POST("", handler.CreatePost)
		posts.PUT("/:id", handler.UpdatePost)
	}
}

func (h *PostHandler) GetPost(c *gin.Context) {
	id := c.Param("id")
	if id == "" || !isValidID(id) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}
	post, err := h.PostUsecase.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch post"})
		return
	}

	if post == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}
	response := domain.PostWithVersion{
		Post:    post.Post,
		Title:   post.Title,
		Content: post.Content,
	}
	c.JSON(http.StatusOK, response)
}

func (h *PostHandler) CreatePost(c *gin.Context) {
	var post domain.CreatePostDTO
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validateCreatePostDTO(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.PostUsecase.Create(&post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})
		return
	}
	response := domain.CreatePostResponseDTO{
		PostID: post.PostID,
		Title:  post.Title,
	}
	c.JSON(http.StatusCreated, response)
}

func (h *PostHandler) UpdatePost(c *gin.Context) {
	id := c.Param("id")
	if id == "" || !isValidID(id) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}
	var post domain.UpdatePostDTO
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validateUpdatePostDTO(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.PostUsecase.Update(id, &post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post"})
		return
	}
	response := domain.UpdatePostResponseDTO{
		Title: post.Title,
	}
	c.JSON(http.StatusOK, response)
}

func validateUpdatePostDTO(request *domain.UpdatePostDTO) error {
	if request.Title == "" && request.Content == "" {
		return errors.New("title or content is required")
	}
	return nil
}

func validateCreatePostDTO(request *domain.CreatePostDTO) error {
	if request.Title == "" || request.Content == "" {
		return errors.New("title and content are required")
	}
	return nil
}

func isValidID(id string) bool {
	return len(id) == 26
}
