package e2e

import (
	"context"
	"database/sql"
	"livoir-blog/internal/app"
	"livoir-blog/mocks"
	"livoir-blog/pkg/auth"
	"livoir-blog/pkg/database"
	"livoir-blog/pkg/jwt"
	"livoir-blog/pkg/logger"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type E2ETestSuite struct {
	suite.Suite
	db                  *sql.DB
	router              *gin.Engine
	pgContainer         testcontainers.Container
	mockOauthRepository *mocks.OAuthRepository
	repoProvider        *app.RepositoryProvider
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

	encryptionKey := "thisisaverysecurekeywith32bytess"

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

	repoProvider, err := app.NewRepositoryProvider(suite.db, oauthConfig, privateKey, publicKey)
	if err != nil {
		suite.T().Fatalf("failed to initialize repository provider: %s", err)
	}
	suite.mockOauthRepository = mocks.NewOAuthRepository(suite.T())
	repoProvider.SetOauthRepository(suite.mockOauthRepository)
	suite.repoProvider = repoProvider

	suite.router, err = app.SetupRouter(suite.db, repoProvider, encryptionKey, time.Duration(60), time.Duration(120))
	if err != nil {
		suite.T().Fatalf("failed to setup router: %s", err)
	}
	suite.insertAdmin("test admin", "admin@example.com")
}

func (suite *E2ETestSuite) TearDownSuite() {
	if suite.db != nil {
		if err := suite.db.Close(); err != nil {
			suite.T().Fatalf("failed to close database: %s", err)
		}
	}
	if suite.pgContainer != nil {
		if err := suite.pgContainer.Terminate(context.Background()); err != nil {
			suite.T().Fatalf("failed to terminate container: %s", err)
		}
	}
}
