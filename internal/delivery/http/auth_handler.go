package http

import (
	"crypto/rand"
	"encoding/base64"
	"livoir-blog/internal/domain"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	OAuthUsecase domain.OAuthUsecase
}

func NewAuthHandler(r *gin.RouterGroup, usecase domain.OAuthUsecase) {
	handler := &AuthHandler{
		OAuthUsecase: usecase,
	}
	r.GET("/google/login", handler.GoogleLogin)
	r.GET("/google/callback", handler.GoogleCallback)
}

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	// Generate state token
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate state token"})
		return
	}
	state := base64.StdEncoding.EncodeToString(b)

	// Store state in session or cookie
	c.SetCookie("state", state, 3600, "/", "", false, true)

	// Redirect to Google's consent page
	url := h.OAuthUsecase.GetRedirectLoginUrl(c.Request.Context(), state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	// Verify state
	state, _ := c.Cookie("state")
	if state != c.Query("state") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid state parameter"})
		return
	}
	code := c.Query("code")
	user, err := h.OAuthUsecase.LoginCallback(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get user info"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"email": user.Email,
		"name":  user.Name,
	})
}
