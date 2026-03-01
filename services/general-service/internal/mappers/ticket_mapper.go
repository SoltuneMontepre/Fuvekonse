package mappers

import (
	"encoding/json"
	"general-service/internal/dto/ticket/responses"
	"general-service/internal/models"
	"general-service/internal/repositories"
)

// MapTicketTierToResponse maps a TicketTier model to a TicketTierResponse DTO
func MapTicketTierToResponse(tier *models.TicketTier) *responses.TicketTierResponse {
	// Parse benefits from JSON string
	var benefits []string
	if tier.Benefits != "" {
		_ = json.Unmarshal([]byte(tier.Benefits), &benefits)
	}

	return &responses.TicketTierResponse{
		ID:          tier.Id,
		TierCode:    tier.TierCode,
		TicketName:  tier.TicketName,
		Description: tier.Description,
		Benefits:    benefits,
		Price:       tier.Price,
		Stock:       tier.Stock,
		IsActive:    tier.IsActive,
	}
}

// MapTicketTiersToResponse maps a slice of TicketTier models to TicketTierResponse DTOs
func MapTicketTiersToResponse(tiers []models.TicketTier) []responses.TicketTierResponse {
	result := make([]responses.TicketTierResponse, len(tiers))
	for i, tier := range tiers {
		result[i] = *MapTicketTierToResponse(&tier)
	}
	return result
}

// MapUserTicketToResponse maps a UserTicket model to a UserTicketResponse DTO
func MapUserTicketToResponse(ticket *models.UserTicket, includeUser bool) *responses.UserTicketResponse {
	response := &responses.UserTicketResponse{
		ID:                    ticket.Id,
		ReferenceCode:         ticket.ReferenceCode,
		Status:                string(ticket.Status),
		TicketNumber:          ticket.TicketNumber,
		ConBadgeName:          ticket.ConBadgeName,
		BadgeImage:            ticket.BadgeImage,
		IsFursuiter:           ticket.IsFursuiter,
		IsFursuitStaff:        ticket.IsFursuitStaff,
		IsCheckedIn:           ticket.IsCheckedIn,
		DenialReason:          ticket.DenialReason,
		CreatedAt:             ticket.CreatedAt,
		ApprovedAt:            ticket.ApprovedAt,
		DeniedAt:              ticket.DeniedAt,
		UpgradedFromTierID:    ticket.UpgradedFromTierID,
		PreviousReferenceCode: ticket.PreviousReferenceCode,
	}

	// Include tier info if available
	if ticket.Ticket.Id != [16]byte{} {
		response.Tier = MapTicketTierToResponse(&ticket.Ticket)
	}

	// Include user info if requested and available (for admin view)
	if includeUser && ticket.User.Id != [16]byte{} {
		response.User = &responses.TicketUserResponse{
			ID:          ticket.User.Id,
			Email:       ticket.User.Email,
			FirstName:   ticket.User.FirstName,
			LastName:    ticket.User.LastName,
			FursonaName: ticket.User.FursonaName,
			DenialCount: ticket.User.DenialCount,
		}
	}

	return response
}

// MapUserTicketsToResponse maps a slice of UserTicket models to UserTicketResponse DTOs
func MapUserTicketsToResponse(tickets []models.UserTicket, includeUser bool) []responses.UserTicketResponse {
	result := make([]responses.UserTicketResponse, len(tickets))
	for i, ticket := range tickets {
		result[i] = *MapUserTicketToResponse(&ticket, includeUser)
	}
	return result
}

// MapTicketStatisticsToResponse maps repository statistics to response DTO
func MapTicketStatisticsToResponse(stats *repositories.TicketStatistics) *responses.TicketStatisticsResponse {
	tierStats := make([]responses.TierStatisticsResponse, len(stats.TierStats))
	for i, ts := range stats.TierStats {
		tierStats[i] = responses.TierStatisticsResponse{
			TierID:     ts.TierID,
			TierCode:   ts.TierCode,
			TierName:   ts.TierName,
			TotalStock: ts.TotalStock,
			Sold:       ts.Sold,
			Available:  ts.Available,
		}
	}

	return &responses.TicketStatisticsResponse{
		TotalTickets:       stats.TotalTickets,
		PendingCount:       stats.PendingCount,
		SelfConfirmedCount: stats.SelfConfirmedCount,
		ApprovedCount:      stats.ApprovedCount,
		DeniedCount:        stats.DeniedCount,
		PendingOver24Hours: stats.PendingOver24Hours,
		TierStats:          tierStats,
	}
}

// MapUserToBlacklistedResponse maps a User model to BlacklistedUserResponse
func MapUserToBlacklistedResponse(user *models.User) *responses.BlacklistedUserResponse {
	return &responses.BlacklistedUserResponse{
		ID:              user.Id,
		Email:           user.Email,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
		FursonaName:     user.FursonaName,
		DenialCount:     user.DenialCount,
		BlacklistedAt:   user.BlacklistedAt,
		BlacklistReason: user.BlacklistReason,
	}
}

// MapUsersToBlacklistedResponse maps a slice of User models to BlacklistedUserResponse DTOs
func MapUsersToBlacklistedResponse(users []models.User) []responses.BlacklistedUserResponse {
	result := make([]responses.BlacklistedUserResponse, len(users))
	for i, user := range users {
		result[i] = *MapUserToBlacklistedResponse(&user)
	}
	return result
}

// MapUpgradeResultToResponse maps a repository UpgradeResult to an UpgradeTicketResponse DTO
func MapUpgradeResultToResponse(result *repositories.UpgradeResult) *responses.UpgradeTicketResponse {
	return &responses.UpgradeTicketResponse{
		Ticket:          MapUserTicketToResponse(result.Ticket, false),
		OldTierPrice:    result.OldTierPrice,
		NewTierPrice:    result.NewTierPrice,
		PriceDifference: result.PriceDifference,
	}
}
