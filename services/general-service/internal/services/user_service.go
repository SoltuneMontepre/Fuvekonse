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

// GetUserByID retrieves a user by their ID and returns public user data without sensitive PII
// Use this for public-facing APIs where user information is exposed
func (s *UserService) GetUserByID(userID string) (*responses.UserResponse, error) {
	user, err := s.repos.User.FindByID(userID)
	if err != nil {
		return nil, err
	}

	// Map model to public response DTO (without sensitive fields)
	userResponse := &responses.UserResponse{
		Id:          user.Id,
		FursonaName: user.FursonaName,
		LastName:    user.LastName,
		FirstName:   user.FirstName,
		Country:     user.Country,
		Avatar:      user.Avatar,
		Role:        user.Role,
		IsVerified:  user.IsVerified,
		CreatedAt:   user.CreatedAt,
		ModifiedAt:  user.ModifiedAt,
	}

	return userResponse, nil
}

// GetUserDetailedByID retrieves a user by their ID and returns detailed user data including sensitive PII
// Use this only for restricted/internal endpoints where users access their own data or admins access user details
func (s *UserService) GetUserDetailedByID(userID string) (*responses.UserDetailedResponse, error) {
	user, err := s.repos.User.FindByID(userID)
	if err != nil {
		return nil, err
	}

	// Map model to detailed response DTO (with sensitive fields)
	userResponse := &responses.UserDetailedResponse{
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
