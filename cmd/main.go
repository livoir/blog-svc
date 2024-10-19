package main

import (
	"fmt"
	"log"
	"strings"

	"livoir-blog/internal/delivery/http"
	"livoir-blog/internal/repository"
	"livoir-blog/internal/usecase"
	"livoir-blog/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	// Initialize Viper
	if err := initConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Read database connection details from Viper
	dbHost := viper.GetString("db.host")
	dbPort := viper.GetString("db.port")
	dbUser := viper.GetString("db.user")
	dbPassword := viper.GetString("db.password")
	dbName := viper.GetString("db.name")

	// Validate that all required configuration is present
	if dbHost == "" || dbPort == "" || dbUser == "" || dbPassword == "" || dbName == "" {
		log.Fatalf("Missing required database configuration")
	}

	db, err := database.NewPostgresConnection(dbHost, dbPort, dbUser, dbPassword, dbName)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.RunMigrations(db, "./migrations"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	postRepo := repository.NewPostRepository(db)
	postUsecase := usecase.NewPostUsecase(postRepo)

	r := gin.Default()
	http.NewPostHandler(r, postUsecase)

	port := viper.GetString("server.port")
	if port == "" {
		log.Fatalf("Server port not specified in configuration")
	}

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
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
			log.Println("No config file found. Ensure all required configuration is set via environment variables.")
		} else {
			return fmt.Errorf("fatal error config file: %s", err)
		}
	}

	return nil
}
