// Package main General Service API
// @title General Service API
// @version 1.0
// @description This is a general service API for managing roles, permissions, and user bans
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8085
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and your JWT token.
package main

import (
	"context"
	_ "general-service/docs"
	"general-service/internal/config"
	"general-service/internal/database"
	"general-service/internal/handlers"
	"general-service/internal/repositories"
	"general-service/internal/services"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var ginLambda *ginadapter.GinLambda

func setupRouter() *gin.Engine {
	config.LoadEnv()

	db, err := database.ConnectWithEnv()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("✅ Database connection established")

	if err := database.InitRedis(); err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
	} else {
		log.Println("✅ Redis connection established")
	}

	repos := repositories.NewRepositories(db)
	svc := services.NewServices(repos)
	h := handlers.NewHandlers(svc)

	router := gin.Default()

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	config.SetupAPIRoutes(router, h, db, database.SetWithExpiration)

	return router
}

func Handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if ginLambda == nil {
		ginLambda = ginadapter.New(setupRouter())
	}
	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		lambda.Start(Handler)
		return
	}

	router := setupRouter()
	port := config.GetEnvOr("PORT", "8080")

	defer database.CloseRedis()

	log.Printf("\n🚀 Server starting on : %s", port)
	log.Printf("📚 Swagger documentation available at: http://localhost:%s/swagger/index.html", port)
	router.Run(":" + port)
}
