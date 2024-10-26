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
}

func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var request domain.CreateCategoryDTO
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.validateCreateCategoryDTO(&request); err != nil {
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

func (h *CategoryHandler) validateCreateCategoryDTO(request *domain.CreateCategoryDTO) error {
	missingFields := []string{}
	if request.Name == "" {
		missingFields = append(missingFields, "name")
	}
	if len(missingFields) > 0 {
		return common.NewCustomError(http.StatusBadRequest, fmt.Sprintf("%s required", strings.Join(missingFields, " and ")))
	}
	return nil
}
