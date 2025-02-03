package http

import (
	"net/http"

	"livoir-blog/internal/domain"

	"github.com/gin-gonic/gin"
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
