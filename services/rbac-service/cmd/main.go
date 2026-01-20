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
	"os"
	"rbac-service/internal/config"
	"rbac-service/internal/database"
	"rbac-service/internal/handlers"
	"rbac-service/internal/repositories"
	"rbac-service/internal/services"

	_ "rbac-service/docs" // Swagger docs - blank import to trigger init()

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var ginLambda *ginadapter.GinLambdaV2

func setupRouter() *gin.Engine {
	config.LoadEnv()

	db, err := database.ConnectWithEnv()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Database connection established")

	if err := database.InitRedis(); err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
	} else {
		log.Println("Redis connection established")
	}

	repos := repositories.NewRepositories(db)
	svc := services.NewServices(repos)
	h := handlers.NewHandlers(svc)

	router := gin.Default()

	// Swagger documentation route
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Check if running in Lambda - if so, API Gateway includes /api/ticket in the path
	isLambda := os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != ""
	if isLambda {
		// In Lambda with API Gateway HTTP API v2, the full path including route prefix is passed
		// Route: /api/ticket/{proxy+} means Lambda receives /api/ticket/...
		ticketGroup := router.Group("/api/ticket")
		config.SetupAPIRoutes(ticketGroup, h, db, database.SetWithExpiration)
	} else {
		// In local development mode, routes start from root
		config.SetupAPIRoutes(router, h, db, database.SetWithExpiration)
		log.Println("Routes configured without prefix for local development")
	}

	return router
}

//Lamdba handler
func Handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	if ginLambda == nil {
		ginLambda = ginadapter.NewV2(setupRouter())
	}
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	// Check if running in AWS Lambda
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		lambda.Start(Handler)
		return
	}

	// Running as HTTP server
	router := setupRouter()
	port := config.GetEnvOr("PORT", "8081")

	defer database.CloseRedis()

	log.Printf("🚀 Server starting on :%s", port)
	log.Printf("📚 Swagger documentation available at: http://localhost:%s/swagger/index.html", port)
	router.Run(":" + port)
}
