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
	"ticket-service/internal/config"
	"ticket-service/internal/database"
	"ticket-service/internal/handlers"
	"ticket-service/internal/repositories"
	"ticket-service/internal/services"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-contrib/cors"
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

	log.Println("✅ Database connection established")

	log.Println("ℹ️  Note: Run 'go run cmd/migrate/main.go' to migrate and seed the database")

	if err := database.InitRedis(); err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
	} else {
		log.Println("✅ Redis connection established")
	}

	repos := repositories.NewRepositories(db)
	// Payment service configuration
	paymentConfig := services.PaymentServiceConfig{
		PayOSClientID:    config.GetEnvOr("PAYOS_CLIENT_ID", "6272354b-095e-45e1-8d9f-0f5f47cecc0f"),
		PayOSAPIKey:      config.GetEnvOr("PAYOS_API_KEY", "faa07056-bc55-49be-a131-0ef24a0e340a"),
		PayOSChecksumKey: config.GetEnvOr("PAYOS_CHECKSUM_KEY", "cdf9415c04208b9ca1b989ea63460808095fb5666af19056a6b1fdea07910033"),
		FrontendURL:      config.GetEnvOr("FRONTEND_URL", "http://localhost:3000"),
		GeneralSvcURL:    config.GetEnvOr("GENERAL_SERVICE_URL", "http://localhost:8080"),
	}

	svc := services.NewServices(repos, paymentConfig)
	h := handlers.NewHandlers(svc)

	router := gin.Default()

	// CORS middleware - allow all origins with credentials
	router.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return true // Allow all origins
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * 3600,
	}))

	// Swagger documentation route
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Check if running in Lambda - if so, API Gateway includes /api/ticket in the path
	isLambda := os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != ""
	if isLambda {
		// In Lambda with API Gateway HTTP API v2, the full path including route prefix is passed
		// Route: /api/ticket/{proxy+} means Lambda receives /api/ticket/...
		ticketGroup := router.Group("/api/ticket")
		config.SetupAPIRoutes(ticketGroup, h, db, database.SetWithExpiration)
		log.Println("Routes configured with /api/ticket prefix for Lambda deployment")
	} else {
		// In local development mode, routes start from root
		config.SetupAPIRoutes(router, h, db, database.SetWithExpiration)
		log.Println("Routes configured without prefix for local development")
	}

	return router
}

// Lamdba handler
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
