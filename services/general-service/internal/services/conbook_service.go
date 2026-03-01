package services

import (
	"context"
	"errors"
	"general-service/internal/dto/conbook/requests"
	"general-service/internal/dto/conbook/responses"
	"general-service/internal/mappers"
	"general-service/internal/models"
	"general-service/internal/repositories"
	"log"

	"github.com/google/uuid"
)

// Re-export sentinel errors from repositories for backward compatibility
var (
	ErrConbookNotFound     = repositories.ErrConbookNotFound
	ErrConbookLimit        = repositories.ErrConbookLimit
	ErrConbookVerified     = repositories.ErrConbookVerified
	ErrUnauthorizedConbook = repositories.ErrUnauthorizedConbook
)

type ConbookService struct {
	repos *repositories.Repositories
}

func NewConbookService(repos *repositories.Repositories) *ConbookService {
	return &ConbookService{repos: repos}
}

// UploadConbook creates a new conbook for a user
// Users can have maximum 10 conbooks
func (s *ConbookService) UploadConbook(ctx context.Context, userIDStr string, req *requests.CreateConbookRequest) (*responses.ConbookResponse, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	// Check conbook limit (max 10)
	count, err := s.repos.Conbook.GetUserConbookCount(ctx, userID)
	if err != nil {
		log.Printf("Error checking conbook count: %v", err)
		return nil, err
	}

	if count >= 10 {
		return nil, ErrConbookLimit
	}

	conbook := &models.ConBookArt{
		Id:          uuid.New(),
		UserId:      userID,
		Title:       req.Title,
		Description: req.Description,
		Handle:      req.Handle,
		ImageUrl:    req.ImageUrl,
		IsVerified:  false,
	}

	created, err := s.repos.Conbook.CreateConbook(ctx, conbook)
	if err != nil {
		log.Printf("Error creating conbook: %v", err)
		return nil, err
	}

	response := mappers.MapConbookToResponse(created)
	return &response, nil
}

// GetUserConbooks retrieves all conbooks for a user
func (s *ConbookService) GetUserConbooks(ctx context.Context, userIDStr string) ([]responses.ConbookResponse, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	conbooks, err := s.repos.Conbook.GetUserConbooks(ctx, userID)
	if err != nil {
		log.Printf("Error retrieving user conbooks: %v", err)
		return nil, err
	}

	return mappers.MapConbooksToResponse(conbooks), nil
}

// GetConbookByID retrieves a single conbook by ID
func (s *ConbookService) GetConbookByID(ctx context.Context, conbookIDStr string) (*responses.ConbookResponse, error) {
	conbookID, err := uuid.Parse(conbookIDStr)
	if err != nil {
		return nil, errors.New("invalid conbook id")
	}

	conbook, err := s.repos.Conbook.GetConbookByID(ctx, conbookID)
	if err != nil {
		return nil, err
	}

	response := mappers.MapConbookToResponse(conbook)
	return &response, nil
}

// EditConbook updates a conbook (only if not verified)
// User can only edit their own conbooks, and only before verification
func (s *ConbookService) EditConbook(ctx context.Context, userIDStr string, conbookIDStr string, req *requests.UpdateConbookRequest) (*responses.ConbookResponse, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	conbookID, err := uuid.Parse(conbookIDStr)
	if err != nil {
		return nil, errors.New("invalid conbook id")
	}

	// Check authorization
	canEdit, err := s.repos.Conbook.CanEditConbook(ctx, userID, conbookID)
	if err != nil {
		return nil, err
	}

	if !canEdit {
		return nil, ErrUnauthorizedConbook
	}

	// Get existing conbook
	existing, err := s.repos.Conbook.GetConbookByID(ctx, conbookID)
	if err != nil {
		return nil, err
	}

	// Apply updates (only update non-nil fields)
	if req.Title != nil {
		existing.Title = *req.Title
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.Handle != nil {
		existing.Handle = *req.Handle
	}
	if req.ImageUrl != nil {
		existing.ImageUrl = *req.ImageUrl
	}

	updated, err := s.repos.Conbook.UpdateConbook(ctx, conbookID, existing)
	if err != nil {
		log.Printf("Error updating conbook: %v", err)
		return nil, err
	}

	response := mappers.MapConbookToResponse(updated)
	return &response, nil
}

// DeleteConbook deletes a conbook (only if not verified)
// User can only delete their own conbooks, and only before verification
func (s *ConbookService) DeleteConbook(ctx context.Context, userIDStr string, conbookIDStr string) error {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return errors.New("invalid user id")
	}

	conbookID, err := uuid.Parse(conbookIDStr)
	if err != nil {
		return errors.New("invalid conbook id")
	}

	// Check authorization
	canEdit, err := s.repos.Conbook.CanEditConbook(ctx, userID, conbookID)
	if err != nil {
		return err
	}

	if !canEdit {
		return ErrUnauthorizedConbook
	}

	if err := s.repos.Conbook.DeleteConbook(ctx, conbookID); err != nil {
		log.Printf("Error deleting conbook: %v", err)
		return err
	}

	return nil
}

// GetUnverifiedConbooks retrieves all conbooks pending verification (staff only)
func (s *ConbookService) GetUnverifiedConbooks(ctx context.Context) ([]responses.ConbookResponse, error) {
	conbooks, err := s.repos.Conbook.GetUnverifiedConbooks(ctx)
	if err != nil {
		log.Printf("Error retrieving unverified conbooks: %v", err)
		return nil, err
	}

	return mappers.MapConbooksToResponse(conbooks), nil
}

// VerifyConbook marks a conbook as verified (staff only)
// After verification, users cannot edit the conbook
func (s *ConbookService) VerifyConbook(ctx context.Context, conbookIDStr string) (*responses.ConbookResponse, error) {
	conbookID, err := uuid.Parse(conbookIDStr)
	if err != nil {
		return nil, errors.New("invalid conbook id")
	}

	// Get the conbook before verification
	conbook, err := s.repos.Conbook.GetConbookByID(ctx, conbookID)
	if err != nil {
		return nil, err
	}

	if conbook.IsVerified {
		return nil, errors.New("conbook is already verified")
	}

	// Verify the conbook
	if err := s.repos.Conbook.VerifyConbook(ctx, conbookID); err != nil {
		log.Printf("Error verifying conbook: %v", err)
		return nil, err
	}

	// Get updated conbook
	updated, err := s.repos.Conbook.GetConbookByID(ctx, conbookID)
	if err != nil {
		return nil, err
	}

	response := mappers.MapConbookToResponse(updated)
	return &response, nil
}
