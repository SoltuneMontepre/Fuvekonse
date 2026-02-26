package mappers

import (
	"general-service/internal/dto/conbook/responses"
	"general-service/internal/models"
)

// MapConbookToResponse maps a ConBookArt model to a ConbookResponse
func MapConbookToResponse(conbook *models.ConBookArt) responses.ConbookResponse {
	return responses.ConbookResponse{
		Id:          conbook.Id,
		UserId:      conbook.UserId,
		Title:       conbook.Title,
		Description: conbook.Description,
		Handle:      conbook.Handle,
		ImageUrl:    conbook.ImageUrl,
		IsVerified:  conbook.IsVerified,
		CreatedAt:   conbook.CreatedAt,
		ModifiedAt:  conbook.ModifiedAt,
	}
}

// MapConbooksToResponse maps a slice of ConBookArt models to ConbookResponse slice
func MapConbooksToResponse(conbooks []models.ConBookArt) []responses.ConbookResponse {
	result := make([]responses.ConbookResponse, len(conbooks))
	for i, conbook := range conbooks {
		result[i] = MapConbookToResponse(&conbook)
	}
	return result
}
