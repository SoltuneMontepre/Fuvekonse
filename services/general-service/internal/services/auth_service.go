package services

import (
	"context"
	"errors"
	"fmt"
	"general-service/internal/common/constants"
	"general-service/internal/common/utils"
	"general-service/internal/dto/auth/requests"
	"general-service/internal/dto/auth/responses"
	"general-service/internal/repositories"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type AuthService struct {
	repos                 *repositories.Repositories
	redisClient           *redis.Client
	loginMaxFail          int
	loginFailBlockMinutes int
}

func NewAuthService(repos *repositories.Repositories, redisClient *redis.Client, loginMaxFail int, loginFailBlockMinutes int) *AuthService {
	return &AuthService{
		repos:                 repos,
		redisClient:           redisClient,
		loginMaxFail:          loginMaxFail,
		loginFailBlockMinutes: loginFailBlockMinutes,
	}
}

// Login authenticates a user and returns tokens

func (s *AuthService) Login(ctx context.Context, req *requests.LoginRequest) (*responses.LoginResponse, error) {

	// Check if user is blocked due to too many failed login attempts
	isBlocked, remainingMinutes, err := utils.IsLoginBlocked(ctx, s.redisClient, req.Email, s.loginMaxFail)
	if err != nil {
		return nil, err
	}
	if isBlocked {
		return nil, fmt.Errorf("%w: please try again in %d minutes", constants.ErrAccountLocked, remainingMinutes+1)
	}

	// Find user by email
	user, err := s.repos.User.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Increment failed attempts even if user doesn't exist to prevent email enumeration
			if err := utils.IncrementLoginFailedAttempts(ctx, s.redisClient, req.Email, s.loginFailBlockMinutes); err != nil {
				// Log the error with context
				fmt.Printf("[ERROR] Failed to increment login attempts for email %s: %v\n", req.Email, err)
				// Security: do not reveal internal error, but fail closed
				return nil, errors.New("internal server error")
			}
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	// Compare password
	if err := utils.ComparePassword(user.Password, req.Password); err != nil {
		// Increment failed login attempts
		if incErr := utils.IncrementLoginFailedAttempts(ctx, s.redisClient, req.Email, s.loginFailBlockMinutes); incErr != nil {
			fmt.Printf("[ERROR] Failed to increment login attempts for email %s: %v\n", req.Email, incErr)
			return nil, errors.New("internal server error")
		}
		return nil, errors.New("invalid email or password")
	}

	// Reset failed login attempts on successful login
	if err := utils.ResetLoginFailedAttempts(ctx, s.redisClient, req.Email); err != nil {
		fmt.Printf("[ERROR] Failed to reset login attempts for email %s: %v\n", req.Email, err)
		// Optionally retry once
		if retryErr := utils.ResetLoginFailedAttempts(ctx, s.redisClient, req.Email); retryErr != nil {
			fmt.Printf("[ERROR] Retry also failed to reset login attempts for email %s: %v\n", req.Email, retryErr)
			// Here you could increment a metric, e.g. metrics.IncResetLoginFailError()
		}
	}

	// Create tokens (convert int role to string for JWT)
	AccessToken, err := utils.CreateAccessToken(user.Id, user.Email, user.FursonaName, user.Role.String())
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

// VerifyOtpAsync verifies OTP and updates user status to Active (IsVerified = true)
func (s *AuthService) VerifyOtp(ctx context.Context, email string, otp string) (bool, error) {
	// Find user by email
	user, err := s.repos.User.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, fmt.Errorf("user not found")
		}
		return false, fmt.Errorf("an error occurred while verifying the OTP: %w", err)
	}

	// Verify OTP and check expiry
	if user.Otp != otp || user.OtpExpiryTime == nil || user.OtpExpiryTime.Before(time.Now()) {
		return false, nil
	}

	// Update user verification status
	user.IsVerified = true
	user.Otp = ""
	user.OtpExpiryTime = nil

	// Update user in database
	if err := s.repos.User.UpdateUserProfile(user); err != nil {
		return false, fmt.Errorf("an error occurred while verifying the OTP: %w", err)
	}

	return true, nil
}

// I write this out if you need to do the regiter func later 🥀
// VerifyOtpAndCompleteRegistrationAsync verifies OTP and completes registration
func (s *AuthService) VerifyOtpAndCompleteRegistration(ctx context.Context, email string, otp string) (bool, error) {
	// Find user by email
	user, err := s.repos.User.FindByEmail(email)
	if err != nil {
		return false, nil
	}

	// Verify OTP and check expiry
	if user.Otp != otp || user.OtpExpiryTime == nil || user.OtpExpiryTime.Before(time.Now()) {
		return false, nil
	}

	// Update user verification status
	user.IsVerified = true
	user.Otp = ""
	user.OtpExpiryTime = nil

	if err := s.repos.User.UpdateUserProfile(user); err != nil {
		return false, err
	}

	return true, nil
}

// ResendOtp resends OTP to user's email
func (s *AuthService) ResendOtp(ctx context.Context, email string, mailService *MailService, fromEmail string) (bool, error) {
	// Find user by email
	user, err := s.repos.User.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, fmt.Errorf("user not found")
		}
		return false, err
	}

	// Check if user is already verified
	if user.IsVerified {
		return false, fmt.Errorf("account is already verified")
	}

	// Generate new OTP
	newOtp, err := utils.GenerateOtp()
	if err != nil {
		return false, fmt.Errorf("failed to generate OTP")
	}

	// 10 mins expire
	expiryTime := time.Now().Add(10 * time.Minute)
	user.Otp = newOtp
	user.OtpExpiryTime = &expiryTime

	// Update user in database
	if err := s.repos.User.UpdateUserProfile(user); err != nil {
		return false, fmt.Errorf("failed to update OTP")
	}

	// Send OTP email
	if err := mailService.SendOtpEmail(ctx, fromEmail, user.Email, newOtp); err != nil {
		return false, fmt.Errorf("failed to send OTP email: %w", err)
	}

	return true, nil
}
