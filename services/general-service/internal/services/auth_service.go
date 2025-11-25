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
	"net/url"
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
				return nil, constants.ErrInternalServer
			}
			return nil, constants.ErrInvalidCredentials
		}
		return nil, err
	}

	// Compare password
	if err := utils.ComparePassword(user.Password, req.Password); err != nil {
		// Increment failed login attempts
		if incErr := utils.IncrementLoginFailedAttempts(ctx, s.redisClient, req.Email, s.loginFailBlockMinutes); incErr != nil {
			fmt.Printf("[ERROR] Failed to increment login attempts for email %s: %v\n", req.Email, incErr)
			return nil, constants.ErrInternalServer
		}
		return nil, constants.ErrInvalidCredentials
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
		return constants.ErrPasswordMismatch
	}

	// Fetch user
	user, err := s.repos.User.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return constants.ErrUserNotFound
		}
		return err
	}

	if err := utils.ComparePassword(user.Password, req.CurrentPassword); err != nil {
		return constants.ErrCurrentPasswordIncorrect
	}

	if req.CurrentPassword == req.NewPassword {
		return constants.ErrSamePassword
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

// ForgotPassword generates a password-reset JWT and emails it to the user.
// frontendURL can be empty; email body includes token itself for Swagger testing.
func (s *AuthService) ForgotPassword(ctx context.Context, email string, mailService *MailService, frontendURL, fromEmail string) error {
	user, err := s.repos.User.FindByEmail(email)
	if err != nil {
		// do not leak existence of email — return nil for not-found
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}

	// create signed reset token
	token, err := utils.CreateForgotPasswordToken(user.Id, user.Email, user.FursonaName, user.Role.String())
	if err != nil {
		return fmt.Errorf("failed to create reset token: %w", err)
	}

	// if frontendURL provided, create link; otherwise include token in email body
	var link string
	if frontendURL != "" {
		u, err := url.Parse(frontendURL)
		if err == nil {
			q := u.Query()
			q.Set("token", token)
			u.RawQuery = q.Encode()
			link = u.String()
		}
	}

	// Compose email body. For Swagger testing we include the raw token as well.
	body := fmt.Sprintf(
		"Hello,\n\nUse the following token to reset your password (expires in %d minutes):\n\n%s\n\n",
		int(utils.GetForgotPasswordTokenExpiry().Minutes()),
		token,
	)
	if link != "" {
		body += fmt.Sprintf("\nOr open the following link:\n\n%s\n\n", link)
	}
	body += "If you did not request this, ignore this message."

	if mailService == nil {
		return fmt.Errorf("mail service not available")
	}

	if err := mailService.SendEmail(ctx, fromEmail, user.Email, "Password Reset Request", body, nil, nil); err != nil {
		return fmt.Errorf("failed to send reset email: %w", err)
	}

	return nil
}

// ResetPasswordWithToken validates reset token and updates the user's password
func (s *AuthService) ResetPasswordWithToken(token string, req *requests.ResetPasswordTokenRequest) error {
	if req.NewPassword != req.ConfirmedPassword {
		return constants.ErrPasswordMismatch
	}

	claims, err := utils.ValidateForgotPasswordToken(token)
	if err != nil {
		return err
	}

	userID := claims.UserID
	user, err := s.repos.User.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return constants.ErrUserNotFound
		}
		return err
	}

	hashed, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return constants.ErrInternalServer
	}

	user.Password = hashed

	if err := s.repos.User.UpdateUserProfile(user); err != nil {
		return constants.ErrInternalServer
	}

	return nil
}
