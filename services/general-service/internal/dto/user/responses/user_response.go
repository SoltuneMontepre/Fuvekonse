package responses

import (
	"time"

	role "general-service/internal/common/constants"

	"github.com/google/uuid"
)

// UserResponse is the public user response DTO without sensitive PII
// This should be used for public APIs where user data is exposed
type UserResponse struct {
	Id          uuid.UUID     `json:"id"`
	FursonaName string        `json:"fursona_name"`
	LastName    string        `json:"last_name"`
	FirstName   string        `json:"first_name"`
	Country     string        `json:"country"`
	Avatar      string        `json:"avatar"`
	Role        role.UserRole `json:"role"`
	IsVerified  bool          `json:"is_verified"`
	CreatedAt   time.Time     `json:"created_at"`
	ModifiedAt  time.Time     `json:"modified_at"`
}

// UserDetailedResponse includes sensitive PII fields
// This should be used only for restricted/internal endpoints where the user
// is accessing their own data or admins are accessing user details
type UserDetailedResponse struct {
	Id              uuid.UUID  `json:"id"`
	FursonaName     string     `json:"fursona_name"`
	LastName        string     `json:"last_name"`
	FirstName       string     `json:"first_name"`
	Country         string     `json:"country"`
	Email           string     `json:"email"`
	Avatar          string     `json:"avatar"`
	Role            role.UserRole `json:"role"`
	IdCard          string     `json:"id_card,omitempty"`
	DateOfBirth     string     `json:"date_of_birth,omitempty"` // "2006-01-02"
	IsVerified      bool       `json:"is_verified"`
	IsDealer        bool       `json:"is_dealer"`
	IsHasTicket     bool       `json:"is_has_ticket"`
	IsBlacklisted   bool       `json:"is_blacklisted"`
	IsBanned        bool       `json:"is_banned"` // Alias for IsBlacklisted so FE can detect ban status
	DenialCount     int        `json:"denial_count"`
	BlacklistedAt   *time.Time `json:"blacklisted_at,omitempty"`
	BlacklistReason string     `json:"blacklist_reason,omitempty"`
	IsDeleted       bool       `json:"is_deleted"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	ModifiedAt      time.Time  `json:"modified_at"`
}

// CountByCountryItem is one entry in the accounts-by-country stats
type CountByCountryItem struct {
	Country string `json:"country"` // Country code (or empty for unknown)
	Count   int    `json:"count"`
}

// CountByCountryResponse is the response for GET /admin/users/statistics/count-by-country
type CountByCountryResponse struct {
	ByCountry []CountByCountryItem `json:"by_country"`
}
