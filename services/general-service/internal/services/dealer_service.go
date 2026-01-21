package services

import (
	"errors"
	"general-service/internal/dto/dealer/requests"
	"general-service/internal/dto/dealer/responses"
	"general-service/internal/mappers"
	"general-service/internal/models"
	"general-service/internal/repositories"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DealerService struct {
	repos *repositories.Repositories
}

func NewDealerService(repos *repositories.Repositories) *DealerService {
	return &DealerService{repos: repos}
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

	// Check if user is already a dealer staff
	existingStaff, err := s.repos.Dealer.FindStaffByUserID(userID)
	if err == nil && existingStaff != nil {
		return nil, errors.New("user is already registered as a dealer staff")
	}

	// Create dealer booth
	booth := &models.DealerBooth{
		Id:              uuid.New(),
		BoothName:       req.BoothName,
		Description:     req.Description,
		PriceSheet:      req.PriceSheet,
		BoothNumber:     "", // Will be assigned later by admin
		IsVerified:      false,
		PaymentVerified: false,
		IsDeleted:       false,
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
