package mappers

import (
	"general-service/internal/dto/dealer/responses"
	"general-service/internal/models"
)

func MapDealerBoothToResponse(booth *models.DealerBooth) *responses.DealerBoothResponse {
	return &responses.DealerBoothResponse{
		Id:          booth.Id,
		BoothName:   booth.BoothName,
		Description: booth.Description,
		BoothNumber: booth.BoothNumber,
		PriceSheet:  booth.PriceSheet,
		IsVerified:  booth.IsVerified,
		CreatedAt:   booth.CreatedAt,
		ModifiedAt:  booth.ModifiedAt,
	}
}

func MapDealerBoothToDetailResponse(booth *models.DealerBooth) *responses.DealerBoothDetailResponse {
	staffs := make([]*responses.DealerStaffResponse, 0, len(booth.Staffs))
	for _, staff := range booth.Staffs {
		if !staff.IsDeleted {
			userName := staff.User.FursonaName
			if userName == "" {
				userName = staff.User.FirstName + " " + staff.User.LastName
			}
			staffs = append(staffs, &responses.DealerStaffResponse{
				Id:         staff.Id,
				UserId:     staff.UserId,
				UserEmail:  staff.User.Email,
				UserName:   userName,
				IsOwner:    staff.IsOwner,
				CreatedAt:  staff.CreatedAt,
				ModifiedAt: staff.ModifiedAt,
			})
		}
	}

	return &responses.DealerBoothDetailResponse{
		Id:          booth.Id,
		BoothName:   booth.BoothName,
		Description: booth.Description,
		BoothNumber: booth.BoothNumber,
		PriceSheet:  booth.PriceSheet,
		IsVerified:  booth.IsVerified,
		CreatedAt:   booth.CreatedAt,
		ModifiedAt:  booth.ModifiedAt,
		Staffs:      staffs,
	}
}
