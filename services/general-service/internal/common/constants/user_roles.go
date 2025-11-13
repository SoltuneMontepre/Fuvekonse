package constants

import (
	"encoding/json"
	"fmt"
)

type UserRole int

const (
	RoleUser   UserRole = 0 // User role
	RoleAdmin  UserRole = 1 // Admin role
	RoleDealer UserRole = 2 // Dealer role
)

// String returns the string representation of the UserRole
func (r UserRole) String() string {
	switch r {
	case RoleUser:
		return "User"
	case RoleAdmin:
		return "Admin"
	case RoleDealer:
		return "Dealer"
	default:
		return fmt.Sprintf("Unknown(%d)", int(r))
	}
}

// MarshalJSON implements json.Marshaler to convert UserRole to string in JSON
func (r UserRole) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

// UnmarshalJSON implements json.Unmarshaler to parse string to UserRole from JSON
func (r *UserRole) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := ParseUserRole(s)
	if err != nil {
		return err
	}
	*r = parsed
	return nil
}

// ParseUserRole parses a string into a UserRole enum value
// Returns an error if the string is not a valid role
func ParseUserRole(s string) (UserRole, error) {
	switch s {
	case "User", "user", "USER":
		return RoleUser, nil
	case "Admin", "admin", "ADMIN":
		return RoleAdmin, nil
	case "Dealer", "dealer", "DEALER":
		return RoleDealer, nil
	default:
		return 0, fmt.Errorf("invalid user role: %s. Valid roles are: User, Admin, Dealer", s)
	}
}

// IsValid checks if the UserRole is a valid enum value
func (r UserRole) IsValid() bool {
	return r >= RoleUser && r <= RoleDealer
}
