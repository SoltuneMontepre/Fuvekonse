package services

import (
	"general-service/internal/dto/user/responses"
	"general-service/internal/mappers"
	"general-service/internal/models"
	"general-service/internal/repositories"

	"gorm.io/gorm"
)

type UserService struct {
	repos *repositories.Repositories
}

func NewUserService(repos *repositories.Repositories) *UserService {
	return &UserService{repos: repos}
}

// isUserDeleted checks if a user is soft-deleted by examining both IsDeleted flag and DeletedAt timestamp
func isUserDeleted(user *models.User) bool {
	return user.IsDeleted || (user.DeletedAt != nil && !user.DeletedAt.IsZero())
}

// GetUserByID retrieves a user by their ID and returns public user data without sensitive PII
// Use this for public-facing APIs where user information is exposed
func (s *UserService) GetUserByID(userID string) (*responses.UserResponse, error) {
	user, err := s.repos.User.FindByID(userID)
	if err != nil {
		return nil, err
	}

	// Additional check: ensure the user is not soft-deleted
	// This provides defense in depth even though the repository filters by is_deleted
	if isUserDeleted(user) {
		return nil, gorm.ErrRecordNotFound
	}

	return mappers.MapUserToResponse(user), nil
}

// GetUserDetailedByID retrieves a user by their ID and returns detailed user data including sensitive PII
// Use this only for restricted/internal endpoints where users access their own data or admins access user details
func (s *UserService) GetUserDetailedByID(userID string) (*responses.UserDetailedResponse, error) {
	user, err := s.repos.User.FindByID(userID)
	if err != nil {
		return nil, err
	}

	// Additional check: ensure the user is not soft-deleted
	// This provides defense in depth even though the repository filters by is_deleted
	if isUserDeleted(user) {
		return nil, gorm.ErrRecordNotFound
	}

	return mappers.MapUserToDetailedResponse(user), nil
}
