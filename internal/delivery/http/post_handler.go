package http

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"livoir-blog/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
)

type PostHandler struct {
	PostUsecase domain.PostUsecase
}

func NewPostHandler(r *gin.RouterGroup, usecase domain.PostUsecase) {
	handler := &PostHandler{
		PostUsecase: usecase,
	}
	r.GET("/:id", handler.GetPost)
	r.POST("", handler.CreatePost)
	r.PUT("/:id", handler.UpdatePost)

}

func (h *PostHandler) GetPost(c *gin.Context) {
	id := c.Param("id")
	if id == "" || !isValidID(id) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}
	post, err := h.PostUsecase.GetByID(c.Request.Context(), id)
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

	response, err := h.PostUsecase.Create(c.Request.Context(), &post)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})
		return
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
	response, err := h.PostUsecase.Update(c.Request.Context(), id, &post)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post"})
		return
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
	missingFields := []string{}
	if request.Title == "" {
		missingFields = append(missingFields, "title")
	}
	if request.Content == "" {
		missingFields = append(missingFields, "content")
	}
	if len(missingFields) > 0 {
		return fmt.Errorf("%s required", strings.Join(missingFields, " and "))
	}
	return nil
}

func isValidID(id string) bool {
	_, err := ulid.Parse(id)
	return err == nil
}
