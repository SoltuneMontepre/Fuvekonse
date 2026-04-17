package services

import (
	"context"
	"errors"
	"fmt"
	"general-service/internal/common/constants"
	"general-service/internal/common/utils"
	"general-service/internal/dto/common"
	"general-service/internal/dto/user/requests"
	"general-service/internal/dto/user/responses"
	"general-service/internal/mappers"
	"general-service/internal/models"
	"general-service/internal/repositories"
	"math"
	"strings"

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

	// Check if user is a dealer: staff of a verified booth, or the one who registered (owner) a booth
	isDealerVerified, err := s.repos.Dealer.CheckUserIsStaffOfVerifiedBooth(userID)
	if err != nil {
		isDealerVerified = false
	}
	isDealerOwner, err := s.repos.Dealer.CheckUserIsOwnerOfBooth(userID)
	if err != nil {
		isDealerOwner = false
	}
	isDealer := isDealerVerified || isDealerOwner

	// Check if user has a ticket: true only when the ticket is approved (not pending/self_confirmed/denied).
	ticket, err := s.repos.Ticket.GetUserTicket(context.Background(), user.Id)
	if err != nil {
		ticket = nil
	}
	isHasTicket := ticket != nil && (ticket.Status == models.TicketStatusApproved || ticket.Status == models.TicketStatusAdminGranted)

	return mappers.MapUserToDetailedResponseWithDealer(user, isDealer, isHasTicket), nil
}

// UpdateProfile updates user profile information
func (s *UserService) UpdateProfile(userID string, req *requests.UpdateProfileRequest) (*responses.UserDetailedResponse, error) {
	// Fetch user
	user, err := s.repos.User.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Update only provided fields (nil = not provided, *string = explicitly set, even if empty)
	if req.FursonaName != nil {
		user.FursonaName = *req.FursonaName
	}
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.Country != nil {
		user.Country = *req.Country
	}
	if req.IdCard != nil {
		user.IdCard = *req.IdCard
	}
	if req.DateOfBirth != nil {
		sTrim := strings.TrimSpace(*req.DateOfBirth)
		if sTrim == "" {
			user.DateOfBirth = nil
		} else {
			dob, err := utils.ParseAndValidateDateOfBirth(sTrim)
			if err != nil {
				return nil, err
			}
			user.DateOfBirth = dob
		}
	}

	// Save updated user
	if err := s.repos.User.UpdateUserProfile(user); err != nil {
		return nil, errors.New("failed to update profile")
	}

	return mappers.MapUserToDetailedResponse(user), nil
}

// UpdateAvatar updates user avatar URL
func (s *UserService) UpdateAvatar(userID string, req *requests.UpdateAvatarRequest) (*responses.UserDetailedResponse, error) {
	// Fetch user
	user, err := s.repos.User.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, constants.ErrUserNotFound
		}
		return nil, err
	}

	// Additional check: ensure the user is not soft-deleted
	// This provides defense in depth even though the repository filters by is_deleted
	if isUserDeleted(user) {
		return nil, gorm.ErrRecordNotFound
	}

	// Update avatar
	user.Avatar = req.Avatar

	// Save updated user
	// Note: Consider implementing optimistic concurrency control using ModifiedAt
	// to prevent race conditions in concurrent update scenarios
	if err := s.repos.User.UpdateUserProfile(user); err != nil {
		return nil, err
	}

	return mappers.MapUserToDetailedResponse(user), nil
}

// GetAllUsers retrieves all users with pagination and optional search (admin only)
func (s *UserService) GetAllUsers(page, pageSize int, search string) ([]*responses.UserDetailedResponse, *common.PaginationMeta, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	users, total, err := s.repos.User.FindAll(page, pageSize, search)
	if err != nil {
		return nil, nil, err
	}

	// Map users to response DTOs
	userResponses := make([]*responses.UserDetailedResponse, len(users))
	for i, user := range users {
		userResponses[i] = mappers.MapUserToDetailedResponse(user)
	}

	// Calculate pagination metadata
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	meta := &common.PaginationMeta{
		CurrentPage: page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
		TotalItems:  total,
	}

	return userResponses, meta, nil
}

// GetUserByIDForAdmin retrieves a user by ID for admin (includes deleted users)
func (s *UserService) GetUserByIDForAdmin(userID string) (*responses.UserDetailedResponse, error) {
	user, err := s.repos.User.FindByIDForAdmin(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, constants.ErrUserNotFound
		}
		return nil, err
	}

	return mappers.MapUserToDetailedResponse(user), nil
}

// UpdateUserByAdmin updates user information by admin
func (s *UserService) UpdateUserByAdmin(userID string, req *requests.AdminUpdateUserRequest) (*responses.UserDetailedResponse, error) {
	// Fetch user (admin can see deleted users)
	user, err := s.repos.User.FindByIDForAdmin(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, constants.ErrUserNotFound
		}
		return nil, err
	}

	// Update only provided fields (nil = not provided, *string = explicitly set, even if empty)
	// Note: Email cannot be changed by admin for security reasons
	if req.FursonaName != nil {
		user.FursonaName = *req.FursonaName
	}
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.Country != nil {
		user.Country = *req.Country
	}
	if req.Avatar != nil {
		user.Avatar = *req.Avatar
	}
	if req.Role != nil {
		user.Role = *req.Role
	}
	if req.IdCard != nil {
		user.IdCard = *req.IdCard
	}
	if req.IsVerified != nil {
		user.IsVerified = *req.IsVerified
	}

	// Save updated user
	if err := s.repos.User.UpdateUser(user); err != nil {
		return nil, errors.New("failed to update user")
	}

	return mappers.MapUserToDetailedResponse(user), nil
}

// GetUserCountByCountry returns counts of non-deleted users grouped by country (admin only)
func (s *UserService) GetUserCountByCountry() (*responses.CountByCountryResponse, error) {
	results, err := s.repos.User.CountByCountry()
	if err != nil {
		return nil, err
	}
	byCountry := make([]responses.CountByCountryItem, len(results))
	for i, r := range results {
		byCountry[i] = responses.CountByCountryItem{
			Country: r.Country,
			Count:   int(r.Count),
		}
	}
	return &responses.CountByCountryResponse{ByCountry: byCountry}, nil
}

// GetUserCountByAgeRange returns counts of non-deleted users grouped into predefined age buckets (admin only).
// Bucket semantics: min is inclusive, max is exclusive.
func (s *UserService) GetUserCountByAgeRange() (*responses.CountByAgeRangeResponse, error) {
	// Fixed buckets (requested pattern: 16-20, 20-25, ...).
	// Keep the final bucket wide so we always "return everything".
	ranges := [][2]int{
		{16, 20},
		{20, 25},
		{25, 30},
		{30, 35},
		{35, 40},
		{40, 45},
		{45, 50},
		{50, 60},
		{60, 100},
		{100, 200},
	}

	results, err := s.repos.User.CountByAgeRanges(ranges)
	if err != nil {
		return nil, err
	}

	// Build lookup map from query results.
	countByLabel := make(map[string]int, len(results))
	for _, r := range results {
		countByLabel[r.Range] = int(r.Count)
	}

	// Always return all predefined buckets (even if count is 0).
	out := make([]responses.CountByAgeRangeItem, 0, len(ranges)+2)
	for _, rg := range ranges {
		minAge := rg[0]
		maxAge := rg[1]
		lookupLabel := fmt.Sprintf("%d-%d", minAge, maxAge)
		label := lookupLabel
		if minAge == 100 {
			label = "100+"
		}
		out = append(out, responses.CountByAgeRangeItem{
			Range: label,
			Min:   minAge,
			Max:   maxAge,
			Count: countByLabel[lookupLabel],
		})
	}

	// Include unknown DOB and other out-of-range buckets too (still "return everything").
	out = append(out,
		responses.CountByAgeRangeItem{Range: "unknown", Min: 0, Max: 0, Count: countByLabel["unknown"]},
		responses.CountByAgeRangeItem{Range: "other", Min: 0, Max: 0, Count: countByLabel["other"]},
	)

	return &responses.CountByAgeRangeResponse{ByAgeRange: out}, nil
}

// DeleteUser soft deletes a user (admin only)
func (s *UserService) DeleteUser(userID string) error {
	// Fetch user (admin can see deleted users)
	user, err := s.repos.User.FindByIDForAdmin(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return constants.ErrUserNotFound
		}
		return err
	}

	// Check if already deleted
	if user.IsDeleted {
		return errors.New("user already deleted")
	}

	// Soft delete
	if err := s.repos.User.DeleteUser(user); err != nil {
		return errors.New("failed to delete user")
	}

	return nil
}

// VerifyUser verifies a user account (admin only)
func (s *UserService) VerifyUser(userID string) (*responses.UserDetailedResponse, error) {
	// Fetch user
	user, err := s.repos.User.FindByIDForAdmin(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, constants.ErrUserNotFound
		}
		return nil, err
	}

	// Verify user
	user.IsVerified = true

	// Save updated user
	if err := s.repos.User.UpdateUser(user); err != nil {
		return nil, errors.New("failed to verify user")
	}

	return mappers.MapUserToDetailedResponse(user), nil
}
