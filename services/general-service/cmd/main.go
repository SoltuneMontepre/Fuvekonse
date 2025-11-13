// Package main General Service API
//	@title			General Service API
//	@version		1.0
//	@description	This is a general service API for managing roles, permissions, and user bans
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT

//	@host		localhost:8085
//	@BasePath	/api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and your JWT token.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"gorm.io/gorm"

	_ "general-service/docs"
	"general-service/internal/config"
	"general-service/internal/database"
	"general-service/internal/handlers"
	"general-service/internal/middlewares"
	"general-service/internal/repositories"
	"general-service/internal/services"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var ginLambda *ginadapter.GinLambdaV2

// validateRequiredEnvVars checks that all required environment variables are set at startup.
// Returns an error if any required variable is missing.
func validateRequiredEnvVars() error {
	requiredVars := []string{
		"DB_HOST",
		"DB_PORT",
		"DB_USER",
		"DB_PASSWORD",
		"DB_NAME",
		"JWT_SECRET",
		// Redis is optional - only warn if missing
	}

	var missing []string
	for _, v := range requiredVars {
		if os.Getenv(v) == "" {
			missing = append(missing, v)
		}
	}

	// Check Redis separately - just warn, don't fail
	if os.Getenv("REDIS_HOST") == "" && os.Getenv("REDIS_URL") == "" {
		log.Println("WARNING: Neither REDIS_HOST nor REDIS_URL is set. Rate limiting may not work properly.")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return nil
}

// setupRedis initializes the Redis connection and logs the result.
// Non-fatal warning if connection fails.
func setupRedis() {
	if err := database.InitRedis(); err != nil {
		log.Printf("WARNING: Failed to connect to Redis: %v", err)
	} else {
		log.Println("Redis connection established successfully")
	}
}

// ...corsMiddleware moved to internal/middlewares/cors.go...

// setupSwagger conditionally enables Swagger documentation based on environment.
// Swagger is only enabled in development and staging environments.
func setupSwagger(router *gin.Engine) {
	env := config.GetEnvOr("ENV", "development")
	if env == "production" {
		log.Println("Swagger documentation disabled in production")
		return
	}

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	log.Println("Swagger documentation enabled")
}

// setupRouter configures and returns a new Gin router with all middleware,
// handlers, and routes configured.
func setupRouter(db *gorm.DB) *gin.Engine {
	// Load environment configuration
	if err := config.LoadEnv(); err != nil {
		log.Printf("WARNING: Error loading .env file: %v", err)
	}

	// Validate required environment variables
	if err := validateRequiredEnvVars(); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Database is already initialized and passed in
	// Set the global DB reference
	database.GlobalDB = db
	
	// Initialize Redis
	setupRedis()

	// Get rate limit configuration
	loginMaxFail := config.GetLoginMaxFail()
	loginFailBlockMinutes := config.GetLoginFailBlockMinutes()

	// Initialize repositories and services
	repos := repositories.NewRepositories(db)
	svc := services.NewServices(repos, database.RedisClient, loginMaxFail, loginFailBlockMinutes)
	h := handlers.NewHandlers(svc)

	// Setup router with middleware
	router := gin.Default()
	allowedOrigins := config.GetEnvOr("CORS_ALLOWED_ORIGINS", "http://localhost:3000")
	
	router.Use(middlewares.CorsMiddleware(allowedOrigins))
	log.Println("CORS middleware configured with allowed origins:", allowedOrigins)

	// Setup Swagger (disabled in production)
	setupSwagger(router)

	// Setup API routes
	config.SetupAPIRoutes(router, h, db, database.SetWithExpiration)

	return router
}

// Handler is the AWS Lambda handler function that proxies requests through Gin.
// It initializes the Gin Lambda adapter on first invocation and reuses it.
var globalDB *gorm.DB

func Handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	if ginLambda == nil {
		// Load environment configuration FIRST
		if err := config.LoadEnv(); err != nil {
			log.Printf("WARNING: Error loading .env file: %v", err)
		}

		// Validate required environment variables
		if err := validateRequiredEnvVars(); err != nil {
			log.Fatalf("Configuration error: %v", err)
		}

		// Setup database connection
		if globalDB == nil {
			var err error
			globalDB, err = database.ConnectWithEnv()
			if err != nil {
				log.Fatalf("Failed to connect to database: %v", err)
			}
			log.Println("Database connection established successfully")
		}

		// Initialize the Gin Lambda adapter
		ginLambda = ginadapter.NewV2(setupRouter(globalDB))
		log.Println("Lambda handler initialized successfully")
	}
	return ginLambda.ProxyWithContext(ctx, req)
}

// main is the entry point for the application.
// It detects whether running in AWS Lambda or as a standalone HTTP server.
func main() {
	// Load environment configuration FIRST
	if err := config.LoadEnv(); err != nil {
		log.Printf("WARNING: Error loading .env file: %v", err)
	}

	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		log.Println("Running in AWS Lambda mode")
		lambda.Start(Handler)
		return
	}

	// Setup database once and reuse
	db, err := database.ConnectWithEnv()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := database.CloseDB(); err != nil {
			log.Printf("Error closing DB: %v", err)
		}
	}()
	defer database.CloseRedis()

	// Setup HTTP server
	router := setupRouter(db)
	port := config.GetEnvOr("PORT", "8080")

	// Start server
	log.Printf("Server starting on port %s", port)

	env := config.GetEnvOr("ENV", "development")
	if env != "production" {
		log.Printf("Swagger documentation: http://localhost:%s/swagger/index.html", port)
	}

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
