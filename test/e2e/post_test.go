package e2e

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"livoir-blog/internal/app"
	"livoir-blog/internal/domain"
	"livoir-blog/pkg/auth"
	"livoir-blog/pkg/database"
	"livoir-blog/pkg/jwt"
	"livoir-blog/pkg/logger"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type E2ETestSuite struct {
	suite.Suite
	db          *sql.DB
	router      *gin.Engine
	pgContainer testcontainers.Container
}

func (suite *E2ETestSuite) SetupSuite() {
	var migrationPath = os.Getenv("LIVOIR_BLOG_MIGRATION_PATH")
	if migrationPath == "" {
		migrationPath = "../../migrations"
	}
	err := logger.Init()
	if err != nil {
		suite.T().Fatalf("failed to initialize logger: %s", err)
	}
	gin.SetMode(gin.TestMode)

	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(30 * time.Second),
	}

	pgContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		suite.T().Fatalf("failed to start container: %s", err)
	}

	suite.pgContainer = pgContainer

	host, err := pgContainer.Host(ctx)
	if err != nil {
		suite.T().Fatalf("failed to get container host: %s", err)
	}

	port, err := pgContainer.MappedPort(ctx, "5432")
	if err != nil {
		suite.T().Fatalf("failed to get container port: %s", err)
	}

	suite.db, err = database.NewPostgresConnection(host, port.Port(), "testuser", "testpass", "testdb")
	if err != nil {
		suite.T().Fatalf("failed to connect to database: %s", err)
	}

	if err := database.RunMigrations(suite.db, migrationPath); err != nil {
		suite.T().Fatalf("failed to run migrations: %s", err)
	}
	oauthConfig := auth.NewGoogleOauthConfig()
	privateKey, publicKey, err := jwt.NewJWT("../../configs/server.key", "../../configs/server.pem")
	if err != nil {
		suite.T().Fatalf("failed to initialize JWT keys: %s", err)
	}
	suite.router, err = app.SetupRouter(suite.db, oauthConfig, privateKey, publicKey)
	if err != nil {
		suite.T().Fatalf("failed to setup router: %s", err)
	}
}

func (suite *E2ETestSuite) TearDownSuite() {
	if suite.db != nil {
		suite.db.Close()
	}
	if suite.pgContainer != nil {
		if err := suite.pgContainer.Terminate(context.Background()); err != nil {
			suite.T().Fatalf("failed to terminate container: %s", err)
		}
	}
}

func (suite *E2ETestSuite) TestCreateAndGetPost() {
	// Test creating a post with potentially unsafe content
	newPost := domain.CreatePostDTO{
		Title:   "Test Post<script>alert('XSS')</script>",
		Content: "This is a <script>alert('XSS')</script>test post content with <b>some bold text</b>",
	}
	jsonValue, err := json.Marshal(newPost)
	assert.NoError(suite.T(), err)
	req, err := http.NewRequest(http.MethodPost, "/posts", bytes.NewBuffer(jsonValue))
	assert.NoError(suite.T(), err)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var createdPost domain.PostResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &createdPost)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), createdPost.PostID)

	// Test getting the created post
	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("/posts/%s", createdPost.PostID), nil)
	assert.NoError(suite.T(), err)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var retrievedPost domain.PostDetailDTO
	err = json.Unmarshal(w.Body.Bytes(), &retrievedPost)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), createdPost.PostID, retrievedPost.Post.ID)
	assert.Equal(suite.T(), "Test Post", retrievedPost.Title)
	assert.Equal(suite.T(), "This is a test post content with <b>some bold text</b>", retrievedPost.Content)
	assert.NotEmpty(suite.T(), retrievedPost.CreatedAt)
	assert.NotEmpty(suite.T(), retrievedPost.UpdatedAt)
	assert.Empty(suite.T(), retrievedPost.CurrentVersionID)
	assert.Equal(suite.T(), retrievedPost.CurrentVersionID, retrievedPost.Post.CurrentVersionID)
	assert.Empty(suite.T(), retrievedPost.Categories)
}

func (suite *E2ETestSuite) TestGetNonExistentPost() {
	req, err := http.NewRequest(http.MethodGet, "/posts/99999999999999999999999999", nil)
	assert.NoError(suite.T(), err)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
	assert.Equal(suite.T(), "Invalid post ID", response["error"])

	req, err = http.NewRequest(http.MethodGet, "/posts/01JAQDCB26N888RY1ZQ4N6N9YN", nil)
	assert.NoError(suite.T(), err)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
	assert.Equal(suite.T(), "The requested post was not found", response["error"])
}

func (suite *E2ETestSuite) TestUpdatePost() {
	newPost := domain.CreatePostDTO{
		Title:   "Test Post",
		Content: "This is a <script>alert('XSS')</script>test post content with <b>some bold text</b>",
	}
	jsonValue, err := json.Marshal(newPost)
	assert.NoError(suite.T(), err)
	req, err := http.NewRequest(http.MethodPost, "/posts", bytes.NewBuffer(jsonValue))
	assert.NoError(suite.T(), err)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var createdPost domain.PostResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &createdPost)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), createdPost.PostID)

	updatedPost := domain.UpdatePostDTO{
		Title:   "Updated Test Post<script>alert('XSS')</script>",
		Content: "This is an updated <script>alert('XSS')</script>test post content with <b>some bold text</b>"}
	jsonValue, err = json.Marshal(updatedPost)
	assert.NoError(suite.T(), err)
	req, err = http.NewRequest("PUT", fmt.Sprintf("/posts/%s", createdPost.PostID), bytes.NewBuffer(jsonValue))
	assert.NoError(suite.T(), err)

	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response domain.PostResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated Test Post", response.Title)

	// Test getting the updated post
	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("/posts/%s", createdPost.PostID), nil)
	assert.NoError(suite.T(), err)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var retrievedPost domain.PostDetailDTO
	err = json.Unmarshal(w.Body.Bytes(), &retrievedPost)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), createdPost.PostID, retrievedPost.Post.ID)
	assert.Equal(suite.T(), "Updated Test Post", retrievedPost.Title)
	assert.Equal(suite.T(), "This is an updated test post content with <b>some bold text</b>", retrievedPost.Content)
	assert.Equal(suite.T(), int64(1), retrievedPost.VersionNumber)
	assert.Equal(suite.T(), retrievedPost.CreatedAt, retrievedPost.UpdatedAt)
	assert.Empty(suite.T(), retrievedPost.Categories)
}

func (suite *E2ETestSuite) TestPublishPost() {
	// Create a new post
	newPost := domain.CreatePostDTO{
		Title:   "Test Publish Post",
		Content: "This is a test post content for publishing",
	}
	jsonValue, err := json.Marshal(newPost)
	assert.NoError(suite.T(), err)
	req, err := http.NewRequest(http.MethodPost, "/posts", bytes.NewBuffer(jsonValue))
	assert.NoError(suite.T(), err)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var createdPost domain.PostResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &createdPost)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), createdPost.PostID)

	// Publish the post
	req, err = http.NewRequest(http.MethodPost, fmt.Sprintf("/posts/%s/publish", createdPost.PostID), nil)
	assert.NoError(suite.T(), err)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var publishedPost domain.PublishResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &publishedPost)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), publishedPost.PublishedAt)
	assert.Equal(suite.T(), "Test Publish Post", publishedPost.Title)
	assert.Equal(suite.T(), "This is a test post content for publishing", publishedPost.Content)

	// Update the post
	updatedPost := domain.UpdatePostDTO{
		Title:   "Updated Test Publish Post",
		Content: "This is an updated test post content for publishing",
	}
	jsonValue, err = json.Marshal(updatedPost)
	assert.NoError(suite.T(), err)
	req, err = http.NewRequest(http.MethodPut, fmt.Sprintf("/posts/%s", createdPost.PostID), bytes.NewBuffer(jsonValue))
	assert.NoError(suite.T(), err)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Get the updated post
	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("/posts/%s", createdPost.PostID), nil)
	assert.NoError(suite.T(), err)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var retrievedPost domain.PostDetailDTO
	err = json.Unmarshal(w.Body.Bytes(), &retrievedPost)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), createdPost.PostID, retrievedPost.Post.ID)
	assert.Equal(suite.T(), "Updated Test Publish Post", retrievedPost.Title)
	assert.Equal(suite.T(), "This is an updated test post content for publishing", retrievedPost.Content)
	assert.Equal(suite.T(), int64(2), retrievedPost.VersionNumber)
	assert.NotEmpty(suite.T(), retrievedPost.CurrentVersionID)
	assert.Empty(suite.T(), retrievedPost.Categories)

	// Publish the post again
	req, err = http.NewRequest(http.MethodPost, fmt.Sprintf("/posts/%s/publish", createdPost.PostID), nil)
	assert.NoError(suite.T(), err)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var republishedPost domain.PublishResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &republishedPost)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), republishedPost.PostID)
	assert.NotEmpty(suite.T(), republishedPost.PublishedAt)
	assert.Equal(suite.T(), "Updated Test Publish Post", republishedPost.Title)
	assert.Equal(suite.T(), "This is an updated test post content for publishing", republishedPost.Content)
}

func (suite *E2ETestSuite) TestDeleteUnpublishedPost() {
	// Create a new post
	newPost := domain.CreatePostDTO{
		Title:   "Test Delete Unpublished Post",
		Content: "This is a test post content for deletion",
	}
	jsonValue, err := json.Marshal(newPost)
	assert.NoError(suite.T(), err)
	req, err := http.NewRequest(http.MethodPost, "/posts", bytes.NewBuffer(jsonValue))
	assert.NoError(suite.T(), err)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var createdPost domain.PostResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &createdPost)
	assert.NoError(suite.T(), err)

	// Delete the unpublished post
	req, err = http.NewRequest(http.MethodDelete, fmt.Sprintf("/posts/%s", createdPost.PostID), nil)
	assert.NoError(suite.T(), err)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Try to get the deleted post
	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("/posts/%s", createdPost.PostID), nil)
	assert.NoError(suite.T(), err)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *E2ETestSuite) TestDeletePublishedPost() {
	// Create a new post
	newPost := domain.CreatePostDTO{
		Title:   "Test Delete Published Post",
		Content: "This is a test post content for deletion after publishing",
	}
	jsonValue, err := json.Marshal(newPost)
	assert.NoError(suite.T(), err)
	req, err := http.NewRequest(http.MethodPost, "/posts", bytes.NewBuffer(jsonValue))
	assert.NoError(suite.T(), err)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var createdPost domain.PostResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &createdPost)
	assert.NoError(suite.T(), err)

	// Publish the post
	req, err = http.NewRequest(http.MethodPost, fmt.Sprintf("/posts/%s/publish", createdPost.PostID), nil)
	assert.NoError(suite.T(), err)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Try to delete the published post
	req, err = http.NewRequest(http.MethodDelete, fmt.Sprintf("/posts/%s", createdPost.PostID), nil)
	assert.NoError(suite.T(), err)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Expect a bad request or forbidden status, as deleting a published post should not be allowed
	assert.Equal(suite.T(), http.StatusConflict, w.Code)

	// Verify the post still exists
	req, err = http.NewRequest(http.MethodGet, fmt.Sprintf("/posts/%s", createdPost.PostID), nil)
	assert.NoError(suite.T(), err)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func TestE2E(t *testing.T) {
	suite.Run(t, new(E2ETestSuite))
}
