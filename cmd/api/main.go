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
	userRepo := repository.NewUserRepository(database.DB)
	authService := services.NewAuthService(userRepo, database.Rdb, cfg)
	authHandler := handlers.NewAuthHandler(authService)

	// Setup Router
	r := gin.Default()

	// Setup Routes
	routes.SetupRoutes(r, authHandler, cfg, database.Rdb)

	// Start Server
	log.Printf("Server starting on port 8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
