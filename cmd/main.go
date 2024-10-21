package main

import (
	"fmt"
	"livoir-blog/internal/app"
	"livoir-blog/pkg/database"
	"livoir-blog/pkg/logger"
	"log"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	if err := logger.Init(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Initialize Viper
	if err := initConfig(); err != nil {
		logger.Log.Fatal("Failed to load config", zap.Error(err))
	}

	// Read database connection details from Viper
	dbHost := viper.GetString("db.host")
	dbPort := viper.GetString("db.port")
	dbUser := viper.GetString("db.user")
	dbPassword := viper.GetString("db.password")
	dbName := viper.GetString("db.name")

	// Validate that all required configuration is present
	if dbHost == "" || dbPort == "" || dbUser == "" || dbPassword == "" || dbName == "" {
		logger.Log.Fatal("Missing required database configuration")
	}

	db, err := database.NewPostgresConnection(dbHost, dbPort, dbUser, dbPassword, dbName)
	if err != nil {
		logger.Log.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Run migrations
	if err := database.RunMigrations(db, "./migrations"); err != nil {
		logger.Log.Fatal("Failed to run migrations", zap.Error(err))
	}

	router := app.SetupRouter(db)

	port := viper.GetString("server.port")
	if port == "" {
		logger.Log.Fatal("Server port not specified in configuration")
	}

	if err := router.Run(":" + port); err != nil {
		logger.Log.Fatal("Failed to run server", zap.Error(err))
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
			return fmt.Errorf("fatal error config file: %s", err)
		}
	}

	return nil
}
