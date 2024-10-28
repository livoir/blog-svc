package http

import (
	"fmt"
	"livoir-blog/internal/domain"
	"livoir-blog/pkg/common"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct {
	CategoryUsecase domain.CategoryUsecase
}

func NewCategoryHandler(r *gin.RouterGroup, usecase domain.CategoryUsecase) {
	handler := &CategoryHandler{
		CategoryUsecase: usecase,
	}
	r.POST("", handler.CreateCategory)
	r.PUT("/:id", handler.UpdateCategory)
	r.POST("/attach", handler.AttachCategoryToPostVersion)
}

func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var request domain.CategoryRequestDTO
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.validateCategoryRequestDTO(&request); err != nil {
		handleError(c, err)
		return
	}
	response, err := h.CategoryUsecase.Create(c.Request.Context(), &request)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, response)
}

func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	var request domain.CategoryRequestDTO
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, ok := h.validateAndGetCategoryID(c)
	if !ok {
		handleError(c, common.NewCustomError(http.StatusBadRequest, "invalid category id"))
		return
	}
	if err := h.validateCategoryRequestDTO(&request); err != nil {
		handleError(c, err)
		return
	}
	response, err := h.CategoryUsecase.Update(c.Request.Context(), id, &request)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, response)
}

func (h *CategoryHandler) AttachCategoryToPostVersion(c *gin.Context) {
	var request domain.AttachCategoryToPostVersionRequestDTO
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.validateAttachCategoryToPostVersionRequestDTO(&request); err != nil {
		handleError(c, err)
		return
	}
	if err := h.CategoryUsecase.AttachToPostVersion(c.Request.Context(), &request); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "category attached to post version successfully"})
}

func (h *CategoryHandler) validateCategoryRequestDTO(request *domain.CategoryRequestDTO) error {
	missingFields := []string{}
	if strings.TrimSpace(request.Name) == "" {
		missingFields = append(missingFields, "name")
	}
	if len(missingFields) > 0 {
		return common.NewCustomError(http.StatusBadRequest, fmt.Sprintf("%s required", strings.Join(missingFields, " and ")))
	}
	return nil
}

func (h *CategoryHandler) validateAndGetCategoryID(c *gin.Context) (string, bool) {
	id := c.Param("id")
	if id == "" || !isValidID(id) {
		return "", false
	}
	return id, true
}

func (h *CategoryHandler) validateAttachCategoryToPostVersionRequestDTO(request *domain.AttachCategoryToPostVersionRequestDTO) error {
	missingFields := []string{}
	if request.PostVersionID == "" || !isValidID(request.PostVersionID) {
		missingFields = append(missingFields, "post version id")
	}
	if len(request.CategoryIDs) == 0 {
		missingFields = append(missingFields, "category ids")
	}
	for _, categoryID := range request.CategoryIDs {
		if !isValidID(categoryID) {
			missingFields = append(missingFields, "category id")
			break
		}
	}
	if len(missingFields) > 0 {
		return common.NewCustomError(http.StatusBadRequest, fmt.Sprintf("%s required", strings.Join(missingFields, " and ")))
	}
	return nil
}
