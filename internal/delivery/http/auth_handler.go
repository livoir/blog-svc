package http

import (
	"crypto/rand"
	"encoding/base64"
	"livoir-blog/internal/domain"
	"livoir-blog/pkg/common"
	"livoir-blog/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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
		handleError(c, common.NewCustomError(http.StatusBadRequest, "Failed to generate state token"))
		return
	}
	state := base64.StdEncoding.EncodeToString(b)

	redirect := c.Query("redirect")
	if !isValidRedirectUrl(redirect) {
		handleError(c, common.NewCustomError(http.StatusBadRequest, "Invalid redirect URL"))
		return
	}

	// Store state in session or cookie
	c.SetCookie("state", state, 3600, "/", "", true, true)
	c.SetCookie("redirect", redirect, 3600, "/", "", true, true)

	// Redirect to Google's consent page
	url := h.OAuthUsecase.GetRedirectLoginUrl(c.Request.Context(), state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	// Verify state
	state, err := c.Cookie("state")
	if err != nil {
		handleError(c, common.NewCustomError(http.StatusUnauthorized, "State parameter is missing"))
		return
	}
	if state != c.Query("state") {
		handleError(c, common.NewCustomError(http.StatusUnauthorized, "Invalid state parameter"))
		return
	}
	// Get redirect
	redirect, err := c.Cookie("redirect")
	if err != nil {
		handleError(c, common.NewCustomError(http.StatusUnauthorized, "Redirect parameter is missing"))
		return
	}
	request := &domain.LoginCallbackRequest{
		Code:      c.Query("code"),
		IpAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}
	user, err := h.OAuthUsecase.LoginCallback(c.Request.Context(), request)
	if err != nil {
		handleError(c, err)
		return
	}
	logger.Log.Info("Sucessfully Logged In", zap.Any("user", user))
	c.SetCookie("state", "", -1, "/", "", true, true)
	c.SetCookie("redirect", "", -1, "/", "", true, true)
	c.SetCookie("access_token", user.AccessToken, 300, "/", "", true, true)
	c.Redirect(http.StatusTemporaryRedirect, redirect)
}
