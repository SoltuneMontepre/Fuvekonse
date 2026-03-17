package database

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

// migrateCleanupExampleDotComUsersAndCountry performs an idempotent data cleanup:
// - Soft-deletes users whose email contains "example.com" (typically seed/test accounts)
// - Normalizes any country value longer than 2 characters to "VN"
func migrateCleanupExampleDotComUsersAndCountry(db *gorm.DB) error {
	if !db.Migrator().HasTable("users") {
		log.Println("users table doesn't exist yet, skipping example.com/country cleanup")
		return nil
	}

	// Soft-delete (not hard delete) to avoid breaking foreign keys or audit trails.
	// Condition is broad by request: "have example.com in it".
	deleteResult := db.Exec(`
		UPDATE users
		SET is_deleted = true,
			deleted_at = NOW()
		WHERE is_deleted = false
		  AND email ILIKE '%example.com%';
	`)
	if deleteResult.Error != nil {
		return fmt.Errorf("failed to soft-delete users with example.com email: %w", deleteResult.Error)
	}
	if deleteResult.RowsAffected > 0 {
		log.Printf("Soft-deleted %d users with email containing example.com", deleteResult.RowsAffected)
	} else {
		log.Println("No users found with email containing example.com (or already deleted)")
	}

	countryResult := db.Exec(`
		UPDATE users
		SET country = 'VN'
		WHERE country IS NOT NULL
		  AND LENGTH(country) > 2;
	`)
	if countryResult.Error != nil {
		return fmt.Errorf("failed to normalize users.country to VN: %w", countryResult.Error)
	}
	if countryResult.RowsAffected > 0 {
		log.Printf("Normalized country to 'VN' for %d users with country length > 2", countryResult.RowsAffected)
	} else {
		log.Println("No users needed country normalization (length > 2)")
	}

	return nil
}

