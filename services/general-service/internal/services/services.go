package services

import (
	"general-service/internal/repositories"

	"github.com/redis/go-redis/v9"
)

type Services struct {
	Auth *AuthService
	User *UserService
	Mail *MailService
}

func NewServices(repos *repositories.Repositories, redisClient *redis.Client, loginMaxFail int, loginFailBlockMinutes int) *Services {
	return &Services{
		Auth: NewAuthService(repos, redisClient, loginMaxFail, loginFailBlockMinutes),
		User: NewUserService(repos),
		Mail: NewMailService(repos),
	}
}
