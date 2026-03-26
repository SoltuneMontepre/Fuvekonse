package database

import (
	"fmt"
	"log"
	"os"
	"strings"

	"general-service/internal/security"

	"gorm.io/gorm"
)

const encryptedValuePrefix = "v1:"

type userPIIRow struct {
	Id          string `gorm:"column:id"`
	FirstName   string `gorm:"column:first_name"`
	LastName    string `gorm:"column:last_name"`
	FursonaName string `gorm:"column:fursona_name"`
	Country     string `gorm:"column:country"`
	IdCard      string `gorm:"column:id_card"`
}

// migrateEncryptExistingUserPII encrypts legacy plaintext PII values already stored in users.
// It is idempotent: values with prefix "v1:" are considered already encrypted.
func migrateEncryptExistingUserPII(db *gorm.DB) error {
	if !db.Migrator().HasTable("users") {
		log.Println("users table doesn't exist yet, skipping user PII encryption migration")
		return nil
	}

	keyB64 := os.Getenv(userPIIKeyEnv)
	key, err := security.DecodeBase64Key(keyB64)
	if err != nil {
		return fmt.Errorf("failed to decode %s for migration: %w", userPIIKeyEnv, err)
	}
	c, err := security.NewAESCipher(key)
	if err != nil {
		return fmt.Errorf("failed to initialize AES cipher for migration: %w", err)
	}

	rawDB := db.Session(&gorm.Session{SkipHooks: true})
	var totalUpdated int64

	err = rawDB.Transaction(func(tx *gorm.DB) error {
		var rows []userPIIRow
		if err := tx.Table("users").Find(&rows).Error; err != nil {
			return fmt.Errorf("failed to load users for PII migration: %w", err)
		}

		for _, row := range rows {
			updates := map[string]any{}

			if value, changed, err := encryptIfPlain(c, row.FirstName); err != nil {
				return err
			} else if changed {
				updates["first_name"] = value
			}

			if value, changed, err := encryptIfPlain(c, row.LastName); err != nil {
				return err
			} else if changed {
				updates["last_name"] = value
			}

			if value, changed, err := encryptIfPlain(c, row.FursonaName); err != nil {
				return err
			} else if changed {
				updates["fursona_name"] = value
			}

			if value, changed, err := encryptIfPlain(c, row.Country); err != nil {
				return err
			} else if changed {
				updates["country"] = value
			}

			if value, changed, err := encryptIfPlain(c, row.IdCard); err != nil {
				return err
			} else if changed {
				updates["id_card"] = value
			}

			if len(updates) == 0 {
				continue
			}

			result := tx.Table("users").Where("id = ?", row.Id).Updates(updates)
			if result.Error != nil {
				return fmt.Errorf("failed to update encrypted PII for user %s: %w", row.Id, result.Error)
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
		log.Printf("Encrypted PII for %d existing users", totalUpdated)
	} else {
		log.Println("No existing users needed PII encryption")
	}
	return nil
}

func encryptIfPlain(c *security.AESCipher, value string) (string, bool, error) {
	if value == "" || strings.HasPrefix(value, encryptedValuePrefix) {
		return value, false, nil
	}
	encrypted, err := c.EncryptString(value)
	if err != nil {
		return "", false, err
	}
	return encrypted, true, nil
}

