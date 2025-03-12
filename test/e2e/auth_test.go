package e2e

import (
	"livoir-blog/internal/domain"
	"livoir-blog/pkg/common"
	"net/http"
	"net/http/httptest"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func (suite *E2ETestSuite) TestGoogleLoginRedirect() {
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
			name:           "Google login without redirect",
			prepareMocks:   func() {},
			redirectUrl:    "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Google login with invalid redirect",
			prepareMocks:   func() {},
			redirectUrl:    "http://localhost:8080",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Google login with valid redirect",
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

func (suite *E2ETestSuite) TestGoogleCallback() {
	t := suite.T()

	testCases := []struct {
		name            string
		cookies         map[string]string
		queryParams     map[string]string
		mock            func()
		expectedStatus  int
		expectedCookies []string
	}{
		{
			name:           "state cookie is missing",
			cookies:        map[string]string{},
			queryParams:    map[string]string{},
			mock:           func() {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "state query is missing",
			cookies:        map[string]string{"state": "state"},
			queryParams:    map[string]string{},
			mock:           func() {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "state cookie and query are different",
			cookies:        map[string]string{"state": "state"},
			queryParams:    map[string]string{"state": "invalid_state"},
			mock:           func() {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "redirect cookie is missing",
			cookies:        map[string]string{"state": "state"},
			queryParams:    map[string]string{"state": "state"},
			mock:           func() {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:        "failed to get logged in user",
			cookies:     map[string]string{"state": "state", "redirect": "http://localhost:8081"},
			queryParams: map[string]string{"state": "state", "code": "code"},
			mock: func() {
				suite.mockOauthRepository.On("GetLoggedInUser", mock.Anything, "code").Return(nil, common.NewCustomError(http.StatusInternalServerError, "failed to get logged in user")).Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:        "email doesn't exist",
			cookies:     map[string]string{"state": "state", "redirect": "http://localhost:8081"},
			queryParams: map[string]string{"state": "state", "code": "code"},
			mock: func() {
				suite.mockOauthRepository.On("GetLoggedInUser", mock.Anything, "code").Return(&domain.OAuthUser{
					ID:            "id",
					Email:         "notexists@example.com",
					VerifiedEmail: true,
					Name:          "name",
					Picture:       "picture",
				}, nil).Once()
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "success",
			cookies:     map[string]string{"state": "state", "redirect": "http://localhost:8081"},
			queryParams: map[string]string{"state": "state", "code": "code"},
			mock: func() {
				suite.mockOauthRepository.On("GetLoggedInUser", mock.Anything, "code").Return(&domain.OAuthUser{
					ID:            "id",
					Email:         "admin@example.com",
					VerifiedEmail: true,
					Name:          "name",
					Picture:       "picture",
				}, nil).Once()
			},
			expectedStatus:  http.StatusTemporaryRedirect,
			expectedCookies: []string{"access_token", "refresh_token", "state", "redirect"},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			tc.mock()
			req, err := http.NewRequest("GET", "/auth/google/callback", nil)
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
