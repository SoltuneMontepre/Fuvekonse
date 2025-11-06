package services

import "general-service/internal/repositories"

type Services struct {
	Auth *AuthService
	User *UserService
}

func NewServices(repos *repositories.Repositories) *Services {
	return &Services{
		Auth: NewAuthService(repos),
		User: NewUserService(repos),
	}
}
