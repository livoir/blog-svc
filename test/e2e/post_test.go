package e2e

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"livoir-blog/internal/app"
	"livoir-blog/internal/domain"
	"livoir-blog/pkg/database"
	"livoir-blog/pkg/logger"
	"net/http"
	"net/http/httptest"
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
	err := logger.Init()
	if err != nil {
		suite.T().Fatalf("failed to initialize logger: %s", err)
	}
	gin.SetMode(gin.TestMode)

	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "postgres:17",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(5 * time.Second),
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

	if err := database.RunMigrations(suite.db, "../../migrations"); err != nil {
		suite.T().Fatalf("failed to run migrations: %s", err)
	}

	suite.router = app.SetupRouter(suite.db)
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
		Title:   "Test Post",
		Content: "This is a <script>alert('XSS')</script>test post content with <b>some bold text</b>",
	}
	jsonValue, _ := json.Marshal(newPost)
	req, _ := http.NewRequest(http.MethodPost, "/posts", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var createdPost domain.CreatePostDTO
	err := json.Unmarshal(w.Body.Bytes(), &createdPost)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), createdPost.PostId)

	// Test getting the created post
	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/posts/%s", createdPost.PostId), nil)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var retrievedPost domain.PostWithVersion
	err = json.Unmarshal(w.Body.Bytes(), &retrievedPost)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), createdPost.PostId, retrievedPost.Post.ID)
	assert.Equal(suite.T(), newPost.Title, retrievedPost.Title)
	assert.Equal(suite.T(), "This is a test post content with <b>some bold text</b>", retrievedPost.Content)
}

func (suite *E2ETestSuite) TestGetNonExistentPost() {
	req, _ := http.NewRequest(http.MethodGet, "/posts/9999", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

func (suite *E2ETestSuite) TestUpdatePost() {
	newPost := domain.CreatePostDTO{
		Title:   "Test Post",
		Content: "This is a <script>alert('XSS')</script>test post content with <b>some bold text</b>",
	}
	jsonValue, _ := json.Marshal(newPost)
	req, _ := http.NewRequest(http.MethodPost, "/posts", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var createdPost domain.CreatePostDTO
	err := json.Unmarshal(w.Body.Bytes(), &createdPost)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), createdPost.PostId)

	updatedPost := domain.UpdatePostDTO{
		Title:   "Updated Test Post",
		Content: "This is an updated <script>alert('XSS')</script>test post content with <b>some bold text</b>"}
	jsonValue, _ = json.Marshal(updatedPost)
	req, _ = http.NewRequest("PUT", fmt.Sprintf("/posts/%s", createdPost.PostId), bytes.NewBuffer(jsonValue))

	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response domain.UpdatePostDTO
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), updatedPost.Title, response.Title)

}

func TestE2E(t *testing.T) {
	suite.Run(t, new(E2ETestSuite))
}
