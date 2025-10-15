package main

import (
	"context"
	"gin/internal/config"
	"gin/internal/database"
	"gin/internal/handlers"
	"gin/internal/repositories"
	"gin/internal/services"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv()

	port := config.GetEnvOr("PORT", "8080")

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
	router.Run(":" + port)
}