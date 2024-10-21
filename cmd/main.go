package main

import (
	"livoir-blog/internal/app"
	"livoir-blog/pkg/database"
	"livoir-blog/pkg/logger"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	if err := logger.Init(); err != nil {
		return
	}
	defer logger.Sync()
	// Initialize Viper
	if err := initConfig(); err != nil {
		logger.Log.Error("Failed to load config", zap.Error(err))
		return
	}
	// Read database connection details from Viper
	dbHost := viper.GetString("db.host")
	dbPort := viper.GetString("db.port")
	dbUser := viper.GetString("db.user")
	dbPassword := viper.GetString("db.password")
	dbName := viper.GetString("db.name")

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

	router, err := app.SetupRouter(db)
	if err != nil {
		logger.Log.Error("Failed to setup router", zap.Error(err))
		return
	}

	port := viper.GetString("server.port")
	if port == "" {
		logger.Log.Error("Server port not specified in configuration")
		return
	}

	if err := router.Run(":" + port); err != nil {
		logger.Log.Error("Failed to run server", zap.Error(err))
		return
	}
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
