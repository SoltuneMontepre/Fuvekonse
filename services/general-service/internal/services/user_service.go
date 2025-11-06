package services

import (
	"general-service/internal/dto/user/responses"
	"general-service/internal/repositories"
)

type UserService struct {
	repos *repositories.Repositories
}

func NewUserService(repos *repositories.Repositories) *UserService {
	return &UserService{repos: repos}
}

// GetUserByID retrieves a user by their ID
func (s *UserService) GetUserByID(userID string) (*responses.UserResponse, error) {
	user, err := s.repos.User.FindByID(userID)
	if err != nil {
		return nil, err
	}

	// Map model to response DTO
	userResponse := &responses.UserResponse{
		Id:               user.Id,
		FursonaName:      user.FursonaName,
		LastName:         user.LastName,
		FirstName:        user.FirstName,
		Country:          user.Country,
		Email:            user.Email,
		Avatar:           user.Avatar,
		Role:             user.Role,
		IdentificationId: user.IdentificationId,
		PassportId:       user.PassportId,
		IsVerified:       user.IsVerified,
		CreatedAt:        user.CreatedAt,
		ModifiedAt:       user.ModifiedAt,
	}

	return userResponse, nil
}
