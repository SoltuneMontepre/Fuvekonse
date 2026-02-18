package constants

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type UserRole int

const (
	RoleUser   UserRole = 0 // User role
	RoleAdmin  UserRole = 1 // Admin role
	RoleDealer UserRole = 2 // Dealer role
	RoleStaff  UserRole = 3 // Staff role
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
	case RoleStaff:
		return "Staff"
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
	case "Staff", "staff", "STAFF":
		return RoleStaff, nil
	default:
		return 0, fmt.Errorf("invalid user role: %s. Valid roles are: User, Admin, Dealer, Staff", s)
	}
}

// IsValid checks if the UserRole is a valid enum value
func (r UserRole) IsValid() bool {
	return r >= RoleUser && r <= RoleStaff
}

// Value implements driver.Valuer to convert UserRole to integer for database storage
func (r UserRole) Value() (driver.Value, error) {
	return int64(r), nil
}

// Scan implements sql.Scanner to convert database integer to UserRole
func (r *UserRole) Scan(value interface{}) error {
	if value == nil {
		*r = RoleUser
		return nil
	}

	switch v := value.(type) {
	case int64:
		*r = UserRole(v)
		return nil
	case int32:
		*r = UserRole(v)
		return nil
	case int16:
		*r = UserRole(v)
		return nil
	case int8:
		*r = UserRole(v)
		return nil
	case int:
		*r = UserRole(v)
		return nil
	case uint64:
		*r = UserRole(v)
		return nil
	case uint32:
		*r = UserRole(v)
		return nil
	case uint16:
		*r = UserRole(v)
		return nil
	case uint8:
		*r = UserRole(v)
		return nil
	case uint:
		*r = UserRole(v)
		return nil
	case []byte:
		// Try to parse as integer first (most common case for PostgreSQL)
		if len(v) > 0 {
			// Try parsing as integer string
			var i int64
			if _, err := fmt.Sscanf(string(v), "%d", &i); err == nil {
				*r = UserRole(i)
				return nil
			}
			// Try parsing as JSON integer
			if err := json.Unmarshal(v, &i); err == nil {
				*r = UserRole(i)
				return nil
			}
			// Handle string representation from database (backward compatibility)
			var s string
			if err := json.Unmarshal(v, &s); err == nil {
				parsed, err := ParseUserRole(s)
				if err != nil {
					return fmt.Errorf("failed to parse role from database: %w", err)
				}
				*r = parsed
				return nil
			}
		}
		return fmt.Errorf("cannot scan %T (%v) into UserRole", value, string(v))
	case string:
		// Try to parse as integer string first (most common case)
		var i int64
		if _, err := fmt.Sscanf(v, "%d", &i); err == nil {
			*r = UserRole(i)
			return nil
		}
		// Handle string representation from database (backward compatibility)
		parsed, err := ParseUserRole(v)
		if err != nil {
			return fmt.Errorf("failed to parse role from database: %w", err)
		}
		*r = parsed
		return nil
	default:
		return fmt.Errorf("cannot scan %T (%v) into UserRole", value, value)
	}
}
