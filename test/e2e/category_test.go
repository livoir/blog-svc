package e2e

import (
	"bytes"
	"encoding/json"
	"livoir-blog/internal/domain"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func (suite *E2ETestSuite) TestCreateCategory() {
	t := suite.T()

	testCases := []struct {
		name           string
		requestBody    domain.CreateCategoryDTO
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "Valid category creation",
			requestBody: domain.CreateCategoryDTO{
				Name: "Test Category",
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"name": "Test Category",
			},
		},
		{
			name:           "Invalid category creation - missing name",
			requestBody:    domain.CreateCategoryDTO{},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "name required",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tc.requestBody)
			req, _ := http.NewRequest("POST", "/categories", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			for key, expectedValue := range tc.expectedBody {
				assert.Equal(t, expectedValue, response[key], "Mismatch for key %s", key)

			}
		})
	}
}
