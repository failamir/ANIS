package routes

import (
	"net/http"

	"auth-service/internal/config"
	"auth-service/internal/handlers"
	"auth-service/internal/middleware"
	"auth-service/internal/models"
	"auth-service/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func SetupRoutes(r *gin.Engine, authHandler *handlers.AuthHandler, cfg *config.Config, rdb *redis.Client) {
	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
		}

		// Protected Route Example
		protected := api.Group("/protected")
		protected.Use(middleware.AuthMiddleware(cfg, rdb))
		{
			protected.GET("/profile", func(c *gin.Context) {
				userId, _ := c.Get("user_id")

				// Fetch user profile from DB (without password)
				var user models.User
				if err := database.DB.Where("id = ?", userId).First(&user).Error; err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"message": "Access granted",
					"user_id": userId,
					"email":   user.Email,
					"name":    user.Name,
				})
			})
		}
	}
}
