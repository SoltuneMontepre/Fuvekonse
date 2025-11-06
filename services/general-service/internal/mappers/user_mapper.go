package mappers

import (
	"general-service/internal/dto/user/responses"
	"general-service/internal/models"
)

// MapUserToResponse maps a User model to a public UserResponse DTO (without sensitive fields)
func MapUserToResponse(user *models.User) *responses.UserResponse {
	return &responses.UserResponse{
		Id:          user.Id,
		FursonaName: user.FursonaName,
		LastName:    user.LastName,
		FirstName:   user.FirstName,
		Country:     user.Country,
		Avatar:      user.Avatar,
		Role:        user.Role,
		IsVerified:  user.IsVerified,
		CreatedAt:   user.CreatedAt,
		ModifiedAt:  user.ModifiedAt,
	}
}

// MapUserToDetailedResponse maps a User model to a detailed UserDetailedResponse DTO (with sensitive fields)
func MapUserToDetailedResponse(user *models.User) *responses.UserDetailedResponse {
	return &responses.UserDetailedResponse{
		Id:               user.Id,
		FursonaName:      user.FursonaName,
		LastName:         user.LastName,
		FirstName:        user.FirstName,
		Country:          user.Country,
		Email:            user.Email,
		Avatar:           user.Avatar,
		Role:             user.Role,
		IdentificationId: user.IdentificationId,
		PassportId:       user.PassportId,
		IsVerified:       user.IsVerified,
		CreatedAt:        user.CreatedAt,
		ModifiedAt:       user.ModifiedAt,
	}
}
