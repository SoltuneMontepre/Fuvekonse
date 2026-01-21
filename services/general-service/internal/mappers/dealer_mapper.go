package mappers

import (
	"general-service/internal/dto/dealer/responses"
	"general-service/internal/models"
)

func MapDealerBoothToResponse(booth *models.DealerBooth) *responses.DealerBoothResponse {
	return &responses.DealerBoothResponse{
		Id:              booth.Id,
		BoothName:       booth.BoothName,
		Description:     booth.Description,
		BoothNumber:     booth.BoothNumber,
		PriceSheet:      booth.PriceSheet,
		IsVerified:      booth.IsVerified,
		PaymentVerified: booth.PaymentVerified,
		CreatedAt:       booth.CreatedAt,
		ModifiedAt:      booth.ModifiedAt,
	}
}
