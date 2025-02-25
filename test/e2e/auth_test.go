package e2e

import (
	"net/http"
	"net/http/httptest"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func (suite *E2ETestSuite) TestGoogleLoginRedirect() {
	t := suite.T()

	viper.Set("server.allowed_redirects", []string{"localhost:8081"})

	testCases := []struct {
		name           string
		redirectUrl    string
		expectedStatus int
	}{
		{
			name:           "Google login without redirect",
			redirectUrl:    "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Google login with invalid redirect",
			redirectUrl:    "http://localhost:8080",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Google login with valid redirect",
			redirectUrl:    "http://localhost:8081",
			expectedStatus: http.StatusTemporaryRedirect,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			req, err := http.NewRequest("GET", "/auth/google/login", nil)
			assert.NoError(t, err)
			if tc.redirectUrl != "" {
				q := req.URL.Query()
				q.Add("redirect", tc.redirectUrl)
				req.URL.RawQuery = q.Encode()
			}
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			if tc.expectedStatus == http.StatusTemporaryRedirect {
				cookies := w.Result().Cookies()
				assert.NotEmpty(t, cookies)
				var stateCookie, redirectCookie bool
				for _, cookie := range cookies {
					if cookie.Name == "state" {
						stateCookie = true
					}
					if cookie.Name == "redirect" {
						redirectCookie = true
					}
				}
				assert.True(t, stateCookie)
				assert.True(t, redirectCookie)
			} else {
				assert.Empty(t, w.Result().Cookies())
			}
		})
	}
}
