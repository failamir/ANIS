package main

import (
	"log"

	"auth-service/internal/config"
	"auth-service/internal/handlers"
	"auth-service/internal/repository"
	"auth-service/internal/routes"
	"auth-service/internal/services"
	"auth-service/pkg/database"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load Config
	cfg := config.LoadConfig()

	// Connect to Database
	database.ConnectPostgres(cfg)
	database.ConnectRedis(cfg)

	// Setup Repository and Services
	// Setup Repository and Services
	userRepo := repository.NewUserRepository(database.DB)
	authRepo := repository.NewAuthRepository(database.Rdb)
	authService := services.NewAuthService(userRepo, authRepo, cfg)
	authHandler := handlers.NewAuthHandler(authService)

	// Setup Router
	r := gin.Default()

	// Setup Routes
	routes.SetupRoutes(r, authHandler, cfg, database.Rdb)

	// Start Server
	port := cfg.AppPort
	if port == "" {
		port = "8888"
	}
	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
