package http

import (
	"errors"
	"fmt"
	"livoir-blog/internal/domain"
	"livoir-blog/pkg/common"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
	"github.com/spf13/viper"
)

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
		case http.StatusUnauthorized:
			c.JSON(http.StatusUnauthorized, gin.H{"error": customErr.Message})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
}

func isValidRedirectUrl(redirect string) bool {
	allowedRedirects := viper.GetStringSlice("server.allowed_redirects")
	parsedUrl, err := url.Parse(redirect)
	if err != nil {
		return false
	}
	for _, allowed := range allowedRedirects {
		if allowed == parsedUrl.Host {
			return true
		}
	}
	return false
}
