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

// ResetPassword allows a logged-in user to change their password
func (s *AuthService) ResetPassword(userID string, req *requests.ResetPasswordRequest) error {
	if req.NewPassword != req.ConfirmedPassword {
		return errors.New("new password and confirm password do not match")
	}

	// Fetch user
	user, err := s.repos.User.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	if err := utils.ComparePassword(user.Password, req.CurrentPassword); err != nil {
		return errors.New("current password is incorrect")
	}

	if req.CurrentPassword == req.NewPassword {
		return errors.New("new password cannot be the same as the old password")
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return errors.New("failed to hash password")
	}

	user.Password = hashedPassword
	if err := s.repos.User.UpdateUserProfile(user); err != nil {
		return errors.New("failed to update password")
	}

	return nil
}
