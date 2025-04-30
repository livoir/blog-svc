package main

import (
	"context"
	"fmt"
	"livoir-blog/internal/app"
	"livoir-blog/pkg/auth"
	"livoir-blog/pkg/cache"
	"livoir-blog/pkg/database"
	"livoir-blog/pkg/jwt"
	"livoir-blog/pkg/logger"
	"livoir-blog/pkg/opentelemetry"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	if err := logger.Init(); err != nil {
		return
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Println("Failed to sync logger", err)
		}
	}()
	// Initialize Viper
	if err := initConfig(); err != nil {
		logger.Log.Error("Failed to load config", zap.Error(err))
		return
	}

	// Init otel
	otelHost := viper.GetString("otel.host")
	shutdownOtel, err := opentelemetry.NewTracerProvider(context.Background(), otelHost)
	if err != nil {
		logger.Log.Error("Failed to initialize OpenTelemetry", zap.Error(err))
		return
	}

	// Read database connection details from Viper
	dbHost := viper.GetString("db.host")
	dbPort := viper.GetString("db.port")
	dbUser := viper.GetString("db.user")
	dbPassword := viper.GetString("db.password")
	dbName := viper.GetString("db.database")
	// Validate that all required configuration is present
	if dbHost == "" || dbPort == "" || dbUser == "" || dbPassword == "" || dbName == "" {
		logger.Log.Error("Missing required database configuration")
		return
	}

	db, err := database.NewPostgresConnection(dbHost, dbPort, dbUser, dbPassword, dbName)
	if err != nil {
		logger.Log.Error("Failed to connect to database", zap.Error(err))
		return
	}
	defer db.Close()

	// Run migrations
	if err := database.RunMigrations(db, "./migrations"); err != nil {
		logger.Log.Error("Failed to run migrations", zap.Error(err))
		return
	}

	// Initialize KeyDB connection
	cacheAddress := viper.GetString("cache.address")
	cacheUsername := viper.GetString("cache.username")
	cachePassword := viper.GetString("cache.password")
	cacheDB := viper.GetInt("cache.database")
	if cacheAddress == "" {
		logger.Log.Error("Missing required cache configuration")
		return
	}
	cache, err := cache.NewKeyDBClient(context.Background(), cacheAddress, cacheUsername, cachePassword, cacheDB)
	if err != nil {
		logger.Log.Error("Failed to connect to cache", zap.Error(err))
		return
	}

	// Initialize OAuth2
	oauthGoogleConfig := auth.NewGoogleOauthConfig()
	oauthDiscordConfig := auth.NewDiscordOauthConfig()
	privateKey, publicKey, err := jwt.NewJWT(viper.GetString("auth.jwt.private_key"), viper.GetString("auth.jwt.public_key"))
	if err != nil {
		logger.Log.Error("Failed to initialize JWT keys", zap.Error(err))
		return
	}
	encryptionKey := viper.GetString("auth.encryption_key")
	if len(encryptionKey) != 16 && len(encryptionKey) != 24 && len(encryptionKey) != 32 {
		logger.Log.Error("Invalid encryption key length", zap.String("key", encryptionKey))
		return
	}

	accessTokenExpiration := viper.GetDuration("auth.jwt.access_token_expiration")
	refreshTokenExpiration := viper.GetDuration("auth.jwt.refresh_token_expiration")
	if accessTokenExpiration <= 0 || refreshTokenExpiration <= 0 {
		logger.Log.Error("Invalid token expiration duration")
		return
	}
	if accessTokenExpiration >= refreshTokenExpiration {
		logger.Log.Error("Access token expiration must be less than refresh token expiration")
		return
	}

	repoProvider, err := app.NewRepositoryProvider(db, cache, oauthGoogleConfig, oauthDiscordConfig, privateKey, publicKey)
	if err != nil {
		logger.Log.Error("Failed to initialize repository provider", zap.Error(err))
		return
	}

	router, err := app.SetupRouter(db, repoProvider, encryptionKey, accessTokenExpiration, refreshTokenExpiration)
	if err != nil {
		logger.Log.Error("Failed to setup router", zap.Error(err))
		return
	}

	port := viper.GetString("server.port")
	if port == "" {
		logger.Log.Error("Server port not specified in configuration")
		return
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 3 * time.Second,
	}

	// Create channel for shutdown signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Error("Failed to start server", zap.Error(err))
		}
	}()

	logger.Log.Info("Server is running on port " + port)

	// Wait for shutdown signal
	<-quit
	logger.Log.Info("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown server gracefully
	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Error("Server forced to shutdown:", zap.Error(err))
	}

	// Close database connection
	if err := db.Close(); err != nil {
		logger.Log.Error("Error closing database connection:", zap.Error(err))
	}

	// Close cache connection
	if err := cache.Close(); err != nil {
		logger.Log.Error("Error closing cache connection:", zap.Error(err))
	}

	// Shutdown OpenTelemetry
	if err := shutdownOtel(context.Background()); err != nil {
		logger.Log.Error("Failed to shutdown OpenTelemetry", zap.Error(err))
		return
	}

	logger.Log.Info("Server exited properly")
}

func initConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")

	viper.AutomaticEnv()
	viper.SetEnvPrefix("LIVOIR_BLOG")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logger.Log.Warn("No config file found. Ensure all required configuration is set via environment variables.")
		} else {
			logger.Log.Error("Failed to load config", zap.Error(err))
		}
		return err
	}
	return nil
}
