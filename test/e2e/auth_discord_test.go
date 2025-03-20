package e2e

import (
	"net/http"
	"net/http/httptest"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func (suite *E2ETestSuite) TestDiscordLoginRedirect() {
	t := suite.T()
	mockGetRedirectLoginUrl := func(state string) {
		suite.mockOauthRepository.On("GetRedirectLoginUrl", mock.Anything, state).
			Return("https://example-oauth.com", nil).
			Once()
	}
	viper.Set("server.allowed_redirects", []string{"localhost:8081"})

	testCases := []struct {
		name           string
		prepareMocks   func()
		redirectUrl    string
		expectedStatus int
	}{
		{
			name:           "Discord login without redirect",
			prepareMocks:   func() {},
			redirectUrl:    "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Discord login with invalid redirect",
			prepareMocks:   func() {},
			redirectUrl:    "http://localhost:8080",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Discord login with valid redirect",
			prepareMocks: func() {
				mockGetRedirectLoginUrl(mock.Anything)
			},
			redirectUrl:    "http://localhost:8081",
			expectedStatus: http.StatusTemporaryRedirect,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.prepareMocks()
			req, err := http.NewRequest("GET", "/auth/discord/login", nil)
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

func (suite *E2ETestSuite) TestDiscordCallback() {
	t := suite.T()

	testCases := []struct {
		name            string
		prepareMocks    func()
		cookies         map[string]string
		queryParams     map[string]string
		expectedCookies []string
		expectedStatus  int
	}{
		{
			name:           "Discord callback without state cookie",
			prepareMocks:   func() {},
			cookies:        map[string]string{},
			queryParams:    map[string]string{},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.prepareMocks()
			req, err := http.NewRequest("GET", "/auth/discord/callback", nil)
			assert.NoError(t, err)

			for key, value := range tc.cookies {
				req.AddCookie(&http.Cookie{Name: key, Value: value})
			}

			q := req.URL.Query()
			for key, value := range tc.queryParams {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()
			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
			for _, cookie := range w.Result().Cookies() {
				assert.Contains(t, tc.expectedCookies, cookie.Name)

			}
		})
	}
}
