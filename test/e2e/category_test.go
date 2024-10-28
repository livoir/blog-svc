package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"livoir-blog/internal/domain"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func (suite *E2ETestSuite) TestCreateCategory() {
	t := suite.T()

	testCases := []struct {
		name           string
		requestBody    domain.CategoryRequestDTO
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "Valid category creation",
			requestBody: domain.CategoryRequestDTO{
				Name: "Test Category",
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"name": "Test Category",
			},
		},
		{
			name:           "Invalid category creation - missing name",
			requestBody:    domain.CategoryRequestDTO{},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "name required",
			},
		},
		{
			name: "Invalid category creation - name already exists",
			requestBody: domain.CategoryRequestDTO{
				Name: "Test Category",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "category name already exists",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonBody, err := json.Marshal(tc.requestBody)
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(jsonBody))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			for key, expectedValue := range tc.expectedBody {
				assert.Equal(t, expectedValue, response[key], "Mismatch for key %s", key)
			}
			if tc.expectedStatus == http.StatusCreated {
				assert.NotEmpty(t, response["created_at"])
				assert.NotEmpty(t, response["updated_at"])
				createdAt, err := time.Parse(time.RFC3339, response["created_at"].(string))
				assert.NoError(t, err)
				updatedAt, err := time.Parse(time.RFC3339, response["updated_at"].(string))
				assert.NoError(t, err)
				assert.Equal(t, createdAt, updatedAt)
			}
		})
	}
}

func (suite *E2ETestSuite) TestUpdateCategory() {
	// Create a category first
	createBody := domain.CategoryRequestDTO{Name: "Original Category"}
	jsonCreateBody, err := json.Marshal(createBody)
	assert.NoError(suite.T(), err)
	createReq, err := http.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(jsonCreateBody))
	assert.NoError(suite.T(), err)
	createReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, createReq)

	var createResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &createResponse)
	assert.NoError(suite.T(), err)

	idValue, ok := createResponse["id"]
	assert.True(suite.T(), ok, "Response should contain 'id'")
	categoryID, ok := idValue.(string)
	assert.True(suite.T(), ok, "'id' should be a string")

	// Create another category for duplicate name test
	anotherCreateBody := domain.CategoryRequestDTO{Name: "Another Category"}
	jsonAnotherCreateBody, err := json.Marshal(anotherCreateBody)
	assert.NoError(suite.T(), err)
	anotherCreateReq, err := http.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(jsonAnotherCreateBody))
	assert.NoError(suite.T(), err)
	anotherCreateReq.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, anotherCreateReq)

	testCases := []struct {
		name           string
		categoryID     string
		requestBody    domain.CategoryRequestDTO
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:       "Invalid category update - missing name",
			categoryID: categoryID,
			requestBody: domain.CategoryRequestDTO{
				Name: "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "name required",
			},
		},
		{
			name:       "Invalid category update - non-existent ID",
			categoryID: "01JB65NB945EQG4R16CG0BVEXK",
			requestBody: domain.CategoryRequestDTO{
				Name: "Updated Category",
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "category not found",
			},
		},
		{
			name:       "Valid category update",
			categoryID: categoryID,
			requestBody: domain.CategoryRequestDTO{
				Name: "Updated Category",
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"id":   categoryID,
				"name": "Updated Category",
			},
		},
		{
			name:       "Invalid category update - name is the same as before",
			categoryID: categoryID,
			requestBody: domain.CategoryRequestDTO{
				Name: "Updated Category",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "name is the same as before",
			},
		},
		{
			name:       "Invalid category update - duplicate name with other category",
			categoryID: categoryID,
			requestBody: domain.CategoryRequestDTO{
				Name: "Another Category",
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "category name already exists",
			},
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			jsonBody, err := json.Marshal(tc.requestBody)
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPut, "/categories/"+tc.categoryID, bytes.NewBuffer(jsonBody))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			for key, expectedValue := range tc.expectedBody {
				assert.Equal(t, expectedValue, response[key], "Mismatch for key %s", key)
			}

			if tc.expectedStatus == http.StatusOK {
				assert.NotEmpty(t, response["created_at"])
				assert.NotEmpty(t, response["updated_at"])
				createdAt, err := time.Parse(time.RFC3339, response["created_at"].(string))
				assert.NoError(t, err)
				updatedAt, err := time.Parse(time.RFC3339, response["updated_at"].(string))
				assert.NoError(t, err)
				assert.True(t, updatedAt.After(createdAt))
			}
		})
	}
}

func (suite *E2ETestSuite) TestAttachCategoryToPostVersion() {
	t := suite.T()

	// Create a post first
	newPost := domain.CreatePostDTO{
		Title:   "Test Post for Category Attachment",
		Content: "This is a test post content",
	}
	jsonValue, err := json.Marshal(newPost)
	assert.NoError(t, err)
	req, err := http.NewRequest(http.MethodPost, "/posts", bytes.NewBuffer(jsonValue))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var createdPost domain.PostResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &createdPost)
	assert.NoError(t, err)

	// Create a category
	category := domain.CategoryRequestDTO{
		Name: "Test Category for Attachment",
	}
	jsonValue, err = json.Marshal(category)
	assert.NoError(t, err)
	req, err = http.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(jsonValue))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var createdCategory map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &createdCategory)
	assert.NoError(t, err)
	categoryID := createdCategory["id"].(string)

	testCases := []struct {
		name           string
		requestBody    domain.AttachCategoryToPostVersionRequestDTO
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "Valid category attachment",
			requestBody: domain.AttachCategoryToPostVersionRequestDTO{
				CategoryIDs:   []string{categoryID},
				PostVersionID: createdPost.PostVersionID,
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"message": "category attached to post version successfully",
			},
		},
		{
			name: "Invalid attachment - non-existent category",
			requestBody: domain.AttachCategoryToPostVersionRequestDTO{
				CategoryIDs:   []string{"01JB9RGJ59B46BA25711PKMYAS"},
				PostVersionID: createdPost.PostVersionID,
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "category not found",
			},
		},
		{
			name: "Invalid attachment - non-existent post version",
			requestBody: domain.AttachCategoryToPostVersionRequestDTO{
				CategoryIDs:   []string{categoryID},
				PostVersionID: "01JB9RGJ59B46BA25711PKMYAS",
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error": "The requested post version was not found",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jsonBody, err := json.Marshal(tc.requestBody)
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, "/categories/attach", bytes.NewBuffer(jsonBody))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			suite.router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			for key, expectedValue := range tc.expectedBody {
				assert.Equal(t, expectedValue, response[key])
			}
		})
	}

	t.Run("Valid category attachment", func(t *testing.T) {
		// Get post to verify category attachment
		req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("/posts/%s", createdPost.PostID), nil)
		assert.NoError(t, err)

		w = httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var postResponse map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &postResponse)
		assert.NoError(t, err)

		categories, ok := postResponse["categories"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, categories, 1)
		categoryResponse := categories[0].(map[string]interface{})
		assert.Equal(t, categoryID, categoryResponse["id"])
	})
}
