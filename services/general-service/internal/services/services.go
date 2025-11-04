package services

import "general-service/internal/repositories"

type Services struct {
}

func NewServices(repos *repositories.Repositories) *Services {
	return &Services{}
}
