// Package main Ticket Service API
// @title Ticket Service API
// @version 1.0
// @description This is a ticket service API for managing roles, permissions, and user bans
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8081
// @BasePath /api/v1

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
package main

import (
	"context"
	"log"
	"time"

	"github.com/SoltuneMontepre/Fuvekonse/tree/main/services/ticket-service/internal/config"
	"github.com/SoltuneMontepre/Fuvekonse/tree/main/services/ticket-service/internal/database"
	"github.com/SoltuneMontepre/Fuvekonse/tree/main/services/ticket-service/internal/handlers"
	"github.com/SoltuneMontepre/Fuvekonse/tree/main/services/ticket-service/internal/repositories"
	"github.com/SoltuneMontepre/Fuvekonse/tree/main/services/ticket-service/internal/services"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/SoltuneMontepre/Fuvekonse/tree/main/services/ticket-service/docs"
)

func main() {
	config.LoadEnv()

	port := config.GetEnvOr("PORT", "8081")

	db, err := database.ConnectWithEnv()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("✅ Database connection established")

	if err := database.InitRedis(); err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
	} else {
		defer database.CloseRedis()
		log.Println("✅ Redis connection established")
	}

	repos := repositories.NewRepositories(db)
	svc := services.NewServices(repos)
	h := handlers.NewHandlers(svc)

	router := gin.Default()

	// Swagger documentation route
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/health/db", func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(500, gin.H{"error": "Database connection error"})
			return
		}
		if err := sqlDB.Ping(); err != nil {
			c.JSON(500, gin.H{"error": "Database ping failed"})
			return
		}
		c.JSON(200, gin.H{"status": "database healthy"})
	})

	router.GET("/health/redis", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := database.SetWithExpiration(ctx, "health_check", "ok", time.Minute); err != nil {
			c.JSON(500, gin.H{"error": "Redis connection failed"})
			return
		}
		c.JSON(200, gin.H{"status": "redis healthy"})
	})

	config.SetupAPIRoutes(router, h)

	log.Printf("🚀 Server starting on :%s", port)
	log.Printf("📚 Swagger documentation available at: http://localhost:%s/swagger/index.html", port)
	router.Run(":" + port)
}