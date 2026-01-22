package services

import (
	"context"
	"errors"
	"fmt"
	"general-service/internal/common/constants"
	"general-service/internal/common/utils"
	timeConstants "general-service/internal/constants"
	"general-service/internal/dto/auth/requests"
	"general-service/internal/dto/auth/responses"
	"general-service/internal/models"
	"general-service/internal/repositories"
	"net/url"
	"time"

	"github.com/google/uuid"
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

// Register creates a new user account and sends OTP verification email
func (s *AuthService) Register(ctx context.Context, req *requests.RegisterRequest, mailService *MailService, fromEmail string) (*responses.RegisterResponse, error) {
	// Validate password match
	if req.Password != req.ConfirmPassword {
		return nil, constants.ErrPasswordMismatch
	}

	// Check if user already exists
	existingUser, err := s.repos.User.FindByEmail(req.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, fmt.Errorf("user with this email already exists")
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Generate OTP
	otp, err := utils.GenerateOtp()
	if err != nil {
		return nil, fmt.Errorf("failed to generate OTP: %w", err)
	}

	// Parse full name into first and last name
	firstName, lastName := parseFullName(req.FullName)

	// Create new user
	newUser := &models.User{
		Id:          uuid.New(),
		FursonaName: req.Nickname,
		FirstName:   firstName,
		LastName:    lastName,
		Email:       req.Email,
		Password:    hashedPassword,
		Country:     req.Country,
		IdCard:      req.IdCard,
		IsVerified:  false,
		Role:        constants.RoleUser,
		CreatedAt:   time.Now(),
		ModifiedAt:  time.Now(),
	}

	// Save user to database
	if err := s.repos.User.Create(newUser); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Store OTP in Redis with expiration
	otpExpiry := timeConstants.GetOTPExpiryDuration()
	if err := utils.StoreOTP(ctx, s.redisClient, newUser.Email, otp, otpExpiry); err != nil {
		// Log error but don't fail registration
		fmt.Printf("[ERROR] Failed to store OTP in Redis for %s: %v\n", newUser.Email, err)
		return &responses.RegisterResponse{
			Message: "Registration successful, but failed to store verification code. Please request a new OTP.",
			Email:   newUser.Email,
		}, nil
	}

	// Send OTP email
	if err := mailService.SendOtpEmail(ctx, fromEmail, newUser.Email, otp); err != nil {
		// Log error but don't fail registration
		fmt.Printf("[ERROR] Failed to send OTP email to %s: %v\n", newUser.Email, err)
		return &responses.RegisterResponse{
			Message: "Registration successful, but failed to send verification email. Please request a new OTP.",
			Email:   newUser.Email,
		}, nil
	}

	return &responses.RegisterResponse{
		Message: "Registration successful. Please check your email for OTP verification.",
		Email:   newUser.Email,
	}, nil
}

// parseFullName splits a full name into first and last name
func parseFullName(fullName string) (firstName, lastName string) {
	// Simple implementation - you can make this more sophisticated
	parts := splitName(fullName)
	if len(parts) == 0 {
		return "", ""
	}
	if len(parts) == 1 {
		return parts[0], ""
	}
	// First part is first name, rest is last name
	firstName = parts[0]
	for i := 1; i < len(parts); i++ {
		if i > 1 {
			lastName += " "
		}
		lastName += parts[i]
	}
	return firstName, lastName
}

// splitName splits a name by spaces
func splitName(name string) []string {
	var parts []string
	current := ""
	for _, char := range name {
		if char == ' ' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
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

	// Verify OTP against Redis and delete if valid
	valid, err := utils.VerifyAndDeleteOTP(ctx, s.redisClient, email, otp)
	if err != nil {
		return false, fmt.Errorf("an error occurred while verifying the OTP: %w", err)
	}

	if !valid {
		return false, nil
	}

	// Update user verification status
	user.IsVerified = true

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

	// Verify OTP against Redis and delete if valid
	valid, err := utils.VerifyAndDeleteOTP(ctx, s.redisClient, email, otp)
	if err != nil {
		return false, err
	}

	if !valid {
		return false, nil
	}

	// Update user verification status
	user.IsVerified = true

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

	// Set OTP expiry time
	expiryTime := timeConstants.GetOTPExpiryDuration()
	if err := utils.StoreOTP(ctx, s.redisClient, email, newOtp, expiryTime); err != nil {
		return false, fmt.Errorf("failed to store OTP: %w", err)
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
