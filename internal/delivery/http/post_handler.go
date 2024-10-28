package http

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"livoir-blog/internal/domain"
	"livoir-blog/pkg/common"

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
	r.POST("/:id/publish", handler.PublishPost)
	r.DELETE("/:id", handler.DeletePostVersion)
}

func (h *PostHandler) validateAndGetPostID(c *gin.Context) (string, bool) {
	id := c.Param("id")
	if id == "" || !isValidID(id) {
		return "", false
	}
	return id, true
}

func (h *PostHandler) GetPost(c *gin.Context) {
	id, ok := h.validateAndGetPostID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}
	post, err := h.PostUsecase.GetByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, post)
}

func (h *PostHandler) CreatePost(c *gin.Context) {
	var post domain.CreatePostDTO
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validateCreatePostDTO(&post); err != nil {
		handleError(c, err)
		return
	}
	response, err := h.PostUsecase.Create(c.Request.Context(), &post)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response)
}

func (h *PostHandler) UpdatePost(c *gin.Context) {
	id, ok := h.validateAndGetPostID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}
	var post domain.UpdatePostDTO
	if err := c.ShouldBindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validateUpdatePostDTO(&post); err != nil {
		handleError(c, err)
		return
	}
	response, err := h.PostUsecase.Update(c.Request.Context(), id, &post)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (h *PostHandler) PublishPost(c *gin.Context) {
	id, ok := h.validateAndGetPostID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}
	response, err := h.PostUsecase.Publish(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (h *PostHandler) DeletePostVersion(c *gin.Context) {
	id, ok := h.validateAndGetPostID(c)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}
	err := h.PostUsecase.DeletePostVersionByPostID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Post version deleted"})
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

func handleError(c *gin.Context, err error) {
	if customErr, ok := err.(*common.CustomError); ok {
		switch customErr.StatusCode {
		case http.StatusNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": customErr.Message})
		case http.StatusForbidden:
			c.JSON(http.StatusForbidden, gin.H{"error": customErr.Message})
		case http.StatusConflict:
			c.JSON(http.StatusConflict, gin.H{"error": customErr.Message})
		case http.StatusBadRequest:
			c.JSON(http.StatusBadRequest, gin.H{"error": customErr.Message})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
}
