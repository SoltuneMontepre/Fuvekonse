package services

import (
	"context"
	"errors"
	"general-service/internal/common/utils"
	"general-service/internal/dto/common"
	"general-service/internal/dto/dealer/requests"
	"general-service/internal/dto/dealer/responses"
	"general-service/internal/mappers"
	"general-service/internal/models"
	"general-service/internal/repositories"
	"log"
	"math"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DealerService struct {
	repos *repositories.Repositories
	mail  *MailService
}

func NewDealerService(repos *repositories.Repositories, mail *MailService) *DealerService {
	return &DealerService{repos: repos, mail: mail}
}

// RegisterDealer creates a new dealer booth and assigns the creator as owner
func (s *DealerService) RegisterDealer(userID string, req *requests.DealerRegisterRequest) (*responses.DealerBoothResponse, error) {
	// Check if user exists
	user, err := s.repos.User.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Check if user has an approved ticket
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	ctx := context.Background()
	userTicket, err := s.repos.Ticket.GetUserTicket(ctx, userUUID)
	if err != nil {
		return nil, errors.New("failed to check user ticket")
	}
	if userTicket == nil {
		return nil, errors.New("user must have a ticket to register as a dealer")
	}
	if userTicket.Status != models.TicketStatusApproved {
		return nil, errors.New("user ticket must be approved to register as a dealer")
	}

	// Check if user is already a dealer staff
	existingStaff, err := s.repos.Dealer.FindStaffByUserID(userID)
	if err == nil && existingStaff != nil {
		return nil, errors.New("user is already registered as a dealer staff")
	}

	// Create dealer booth
	booth := &models.DealerBooth{
		Id:          uuid.New(),
		BoothName:   req.BoothName,
		Description: req.Description,
		PriceSheet:  req.PriceSheet,
		BoothNumber: "", // Will be assigned later by admin
		IsVerified:  false,
		IsDeleted:   false,
	}

	// Create dealer staff with user as owner
	staff := &models.UserDealerStaff{
		Id:        uuid.New(),
		UserId:    user.Id,
		IsOwner:   true,
		IsDeleted: false,
	}

	// Create both in a transaction
	if err := s.repos.Dealer.CreateBoothWithStaff(booth, staff); err != nil {
		return nil, errors.New("failed to register dealer")
	}

	return mappers.MapDealerBoothToResponse(booth), nil
}

// GetAllDealersForAdmin retrieves all dealer booths with pagination and filters (admin only)
func (s *DealerService) GetAllDealersForAdmin(page, pageSize int, isVerified *bool) ([]*responses.DealerBoothDetailResponse, *common.PaginationMeta, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	booths, total, err := s.repos.Dealer.FindAllBooths(page, pageSize, isVerified)
	if err != nil {
		return nil, nil, err
	}

	// Map booths to response DTOs
	boothResponses := make([]*responses.DealerBoothDetailResponse, len(booths))
	for i, booth := range booths {
		boothResponses[i] = mappers.MapDealerBoothToDetailResponse(booth)
	}

	// Calculate pagination metadata
	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	meta := &common.PaginationMeta{
		CurrentPage: page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
		TotalItems:  total,
	}

	return boothResponses, meta, nil
}

// GetDealerByIDForAdmin retrieves a dealer booth by ID with staff information (admin only)
func (s *DealerService) GetDealerByIDForAdmin(boothID string) (*responses.DealerBoothDetailResponse, error) {
	booth, err := s.repos.Dealer.FindBoothByIDWithStaffs(boothID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("dealer booth not found")
		}
		return nil, err
	}

	return mappers.MapDealerBoothToDetailResponse(booth), nil
}

// VerifyDealer verifies a dealer booth and generates a unique booth code. Sends an email to the booth owner with dealer den (booth) information.
func (s *DealerService) VerifyDealer(ctx context.Context, boothID string, fromEmail string) (*responses.DealerBoothDetailResponse, error) {
	// Check if booth exists
	booth, err := s.repos.Dealer.FindBoothByID(boothID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("dealer booth not found")
		}
		return nil, err
	}

	// Check if already verified
	if booth.IsVerified {
		return nil, errors.New("dealer booth is already verified")
	}

	// Generate unique booth code
	var boothNumber string
	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		code, err := utils.GenerateBoothCode()
		if err != nil {
			return nil, errors.New("failed to generate booth code")
		}

		// Check if code already exists
		exists, err := s.repos.Dealer.CheckBoothNumberExists(code)
		if err != nil {
			return nil, errors.New("failed to check booth code availability")
		}

		if !exists {
			boothNumber = code
			break
		}
	}

	if boothNumber == "" {
		return nil, errors.New("failed to generate unique booth code after multiple attempts")
	}

	// Verify booth with generated code
	verifiedBooth, err := s.repos.Dealer.VerifyBooth(boothID, boothNumber)
	if err != nil {
		return nil, errors.New("failed to verify dealer booth")
	}

	// Reload booth with staff information
	boothWithStaffs, err := s.repos.Dealer.FindBoothByIDWithStaffs(boothID)
	if err != nil {
		// Fallback to verified booth without staffs
		return mappers.MapDealerBoothToDetailResponse(verifiedBooth), nil
	}

	// Send approval email to the booth owner with dealer den information
	if fromEmail != "" && s.mail != nil {
		for i := range boothWithStaffs.Staffs {
			staff := &boothWithStaffs.Staffs[i]
			if staff.IsOwner && !staff.IsDeleted && staff.User.Email != "" {
				if err := s.mail.SendDealerApprovedEmail(ctx, fromEmail, staff.User.Email, boothWithStaffs.BoothName, boothNumber); err != nil {
					log.Printf("Failed to send dealer approved email to %s: %v", staff.User.Email, err)
				}
				break
			}
		}
	}

	return mappers.MapDealerBoothToDetailResponse(boothWithStaffs), nil
}

// JoinDealerBooth allows a user to join a dealer booth using a booth code
func (s *DealerService) JoinDealerBooth(userID string, boothCode string) (*responses.DealerBoothDetailResponse, error) {
	// Check if user exists
	user, err := s.repos.User.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Check if user has an approved ticket
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	ctx := context.Background()
	userTicket, err := s.repos.Ticket.GetUserTicket(ctx, userUUID)
	if err != nil {
		return nil, errors.New("failed to check user ticket")
	}
	if userTicket == nil {
		return nil, errors.New("user must have a ticket to join a dealer booth")
	}
	if userTicket.Status != models.TicketStatusApproved {
		return nil, errors.New("user ticket must be approved to join a dealer booth")
	}

	// Check if user is already a staff member
	isStaff, err := s.repos.Dealer.CheckUserIsStaff(userID)
	if err != nil {
		return nil, errors.New("failed to check user staff status")
	}
	if isStaff {
		return nil, errors.New("user is already a staff member of a dealer booth")
	}

	// Find booth by code
	booth, err := s.repos.Dealer.FindBoothByCode(boothCode)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("dealer booth not found with provided code")
		}
		return nil, err
	}

	// Check if booth is verified
	if !booth.IsVerified {
		return nil, errors.New("dealer booth must be verified before accepting staff")
	}

	// Create staff member (not as owner)
	staff := &models.UserDealerStaff{
		Id:        uuid.New(),
		UserId:    user.Id,
		BoothId:   booth.Id,
		IsOwner:   false,
		IsDeleted: false,
	}

	if err := s.repos.Dealer.CreateStaff(staff); err != nil {
		return nil, errors.New("failed to join dealer booth")
	}

	// Reload booth with staff information
	boothWithStaffs, err := s.repos.Dealer.FindBoothByIDWithStaffs(booth.Id.String())
	if err != nil {
		return nil, errors.New("failed to retrieve dealer booth information")
	}

	return mappers.MapDealerBoothToDetailResponse(boothWithStaffs), nil
}

// RemoveStaffFromBooth allows a booth owner to remove a staff member
func (s *DealerService) RemoveStaffFromBooth(ownerUserID string, staffUserID string) (*responses.DealerBoothDetailResponse, error) {
	// Check if owner exists
	_, err := s.repos.User.FindByID(ownerUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Get the booth that the owner is part of
	ownerBooth, err := s.repos.Dealer.GetBoothByStaffUserID(ownerUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("you are not a staff member of any dealer booth")
		}
		return nil, err
	}

	// Check if the requesting user is the owner of the booth
	ownerStaff, err := s.repos.Dealer.FindStaffByUserAndBoothID(ownerUserID, ownerBooth.Id.String())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("you are not a staff member of this booth")
		}
		return nil, err
	}

	if !ownerStaff.IsOwner {
		return nil, errors.New("only booth owners can remove staff members")
	}

	// Prevent owner from removing themselves
	if ownerUserID == staffUserID {
		return nil, errors.New("booth owner cannot remove themselves")
	}

	// Find the staff member to remove
	staffToRemove, err := s.repos.Dealer.FindStaffByUserAndBoothID(staffUserID, ownerBooth.Id.String())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("staff member not found in this booth")
		}
		return nil, err
	}

	// Prevent removing another owner (if multiple owners are allowed in the future)
	if staffToRemove.IsOwner {
		return nil, errors.New("cannot remove booth owner")
	}

	// Remove the staff member
	if err := s.repos.Dealer.RemoveStaff(staffToRemove.Id.String()); err != nil {
		return nil, errors.New("failed to remove staff member")
	}

	// Reload booth with updated staff information
	boothWithStaffs, err := s.repos.Dealer.FindBoothByIDWithStaffs(ownerBooth.Id.String())
	if err != nil {
		return nil, errors.New("failed to retrieve updated booth information")
	}

	return mappers.MapDealerBoothToDetailResponse(boothWithStaffs), nil
}

// GetMyDealer retrieves the dealer booth for the current user
func (s *DealerService) GetMyDealer(userID string) (*responses.DealerBoothDetailResponse, error) {
	// Get the booth that the user is a staff member of
	booth, err := s.repos.Dealer.GetBoothByStaffUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user is not a staff member of any dealer booth")
		}
		return nil, err
	}

	// Get booth with staff information
	boothWithStaffs, err := s.repos.Dealer.FindBoothByIDWithStaffs(booth.Id.String())
	if err != nil {
		return nil, errors.New("failed to retrieve dealer booth information")
	}

	return mappers.MapDealerBoothToDetailResponse(boothWithStaffs), nil
}
