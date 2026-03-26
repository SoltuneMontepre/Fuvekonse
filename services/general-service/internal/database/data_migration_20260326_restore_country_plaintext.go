package database

import (
	"fmt"
	"log"
	"os"
	"strings"

	"general-service/internal/security"

	"gorm.io/gorm"
)

type userCountryRow struct {
	Id      string `gorm:"column:id"`
	Country string `gorm:"column:country"`
}

// migrateRestoreCountryPlaintext decrypts users.country back to plaintext.
// It is safe to run multiple times: only values prefixed with "v1:" are decrypted.
func migrateRestoreCountryPlaintext(db *gorm.DB) error {
	if !db.Migrator().HasTable("users") {
		log.Println("users table doesn't exist yet, skipping country restore migration")
		return nil
	}

	keyB64 := os.Getenv(userPIIKeyEnv)
	key, err := security.DecodeBase64Key(keyB64)
	if err != nil {
		return fmt.Errorf("failed to decode %s for country restore migration: %w", userPIIKeyEnv, err)
	}
	c, err := security.NewAESCipher(key)
	if err != nil {
		return fmt.Errorf("failed to initialize AES cipher for country restore migration: %w", err)
	}

	rawDB := db.Session(&gorm.Session{SkipHooks: true})
	var totalUpdated int64

	err = rawDB.Transaction(func(tx *gorm.DB) error {
		var rows []userCountryRow
		if err := tx.Table("users").Find(&rows).Error; err != nil {
			return fmt.Errorf("failed to load users for country restore migration: %w", err)
		}

		for _, row := range rows {
			if row.Country == "" || !strings.HasPrefix(row.Country, encryptedValuePrefix) {
				continue
			}

			plaintext, err := c.DecryptString(row.Country)
			if err != nil {
				return fmt.Errorf("failed to decrypt country for user %s: %w", row.Id, err)
			}

			result := tx.Table("users").Where("id = ?", row.Id).Update("country", plaintext)
			if result.Error != nil {
				return fmt.Errorf("failed to restore country for user %s: %w", row.Id, result.Error)
			}
			if result.RowsAffected > 0 {
				totalUpdated += result.RowsAffected
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	if totalUpdated > 0 {
		log.Printf("Restored plaintext country for %d users", totalUpdated)
	} else {
		log.Println("No users needed country plaintext restore")
	}
	return nil
}

