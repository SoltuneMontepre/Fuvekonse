package services

import (
	"errors"
	"general-service/internal/common/utils"
	"general-service/internal/dto/auth/requests"
	"general-service/internal/dto/auth/responses"
	"general-service/internal/repositories"

	"gorm.io/gorm"
)

type AuthService struct {
	repos *repositories.Repositories
}

func NewAuthService(repos *repositories.Repositories) *AuthService {
	return &AuthService{repos: repos}
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(req *requests.LoginRequest) (*responses.LoginResponse, error) {
	// Find user by email
	user, err := s.repos.User.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	// Compare password
	if err := utils.ComparePassword(user.Password, req.Password); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Create tokens
	AccessToken, err := utils.CreateAccessToken(user.Id, user.Email, user.FursonaName, string(user.Role))
	if err != nil {
		return nil, err
	}

	// Build response
	response := &responses.LoginResponse{
		AccessToken: AccessToken,
	}

	return response, nil
}
