package database

import (
	"fmt"
	"log"
	"os"
	"strings"

	"general-service/internal/security"

	"gorm.io/gorm"
)

type userFursonaNameRow struct {
	Id          string `gorm:"column:id"`
	FursonaName string `gorm:"column:fursona_name"`
}

// migrateRestoreFursonaNamePlaintext decrypts users.fursona_name back to plaintext.
// It is safe to run multiple times: only values prefixed with "v1:" are decrypted.
func migrateRestoreFursonaNamePlaintext(db *gorm.DB) error {
	if !db.Migrator().HasTable("users") {
		log.Println("users table doesn't exist yet, skipping fursona_name restore migration")
		return nil
	}

	keyB64 := os.Getenv(userPIIKeyEnv)
	key, err := security.DecodeBase64Key(keyB64)
	if err != nil {
		return fmt.Errorf("failed to decode %s for fursona_name restore migration: %w", userPIIKeyEnv, err)
	}
	c, err := security.NewAESCipher(key)
	if err != nil {
		return fmt.Errorf("failed to initialize AES cipher for fursona_name restore migration: %w", err)
	}

	rawDB := db.Session(&gorm.Session{SkipHooks: true})
	var totalUpdated int64

	err = rawDB.Transaction(func(tx *gorm.DB) error {
		var rows []userFursonaNameRow
		if err := tx.Table("users").Find(&rows).Error; err != nil {
			return fmt.Errorf("failed to load users for fursona_name restore migration: %w", err)
		}

		for _, row := range rows {
			if row.FursonaName == "" || !strings.HasPrefix(row.FursonaName, encryptedValuePrefix) {
				continue
			}

			plaintext, err := c.DecryptString(row.FursonaName)
			if err != nil {
				return fmt.Errorf("failed to decrypt fursona_name for user %s: %w", row.Id, err)
			}

			result := tx.Table("users").Where("id = ?", row.Id).Update("fursona_name", plaintext)
			if result.Error != nil {
				return fmt.Errorf("failed to restore fursona_name for user %s: %w", row.Id, result.Error)
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
		log.Printf("Restored plaintext fursona_name for %d users", totalUpdated)
	} else {
		log.Println("No users needed fursona_name plaintext restore")
	}
	return nil
}
