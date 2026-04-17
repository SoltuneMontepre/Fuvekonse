package database

import (
	"fmt"
	"os"
	"reflect"

	"general-service/internal/models"
	"general-service/internal/security"

	"gorm.io/gorm"
)

const userPIIKeyEnv = "USER_PII_AES_KEY"

// RegisterUserPIIEncryption installs GORM callbacks that transparently encrypt/decrypt
// selected PII fields for models.User.
func RegisterUserPIIEncryption(db *gorm.DB) error {
	keyB64 := os.Getenv(userPIIKeyEnv)
	key, err := security.DecodeBase64Key(keyB64)
	if err != nil {
		return fmt.Errorf("%s is invalid: %w", userPIIKeyEnv, err)
	}
	c, err := security.NewAESCipher(key)
	if err != nil {
		return fmt.Errorf("%s is invalid: %w", userPIIKeyEnv, err)
	}

	encryptDest := func(tx *gorm.DB) {
		if tx.Statement == nil {
			return
		}
		_ = walkAndApplyUsers(tx.Statement.Dest, func(u *models.User) error {
			return encryptUserPII(c, u)
		})
	}

	decryptDest := func(tx *gorm.DB) {
		if tx.Statement == nil {
			return
		}
		_ = walkAndApplyUsers(tx.Statement.Dest, func(u *models.User) error {
			return decryptUserPII(c, u)
		})
	}

	// Create / Update: encrypt before writing.
	db.Callback().Create().Before("gorm:create").Register("user_pii_encrypt_create", encryptDest)
	db.Callback().Update().Before("gorm:update").Register("user_pii_encrypt_update", encryptDest)

	// Query: decrypt after scanning into destination.
	db.Callback().Query().After("gorm:query").Register("user_pii_decrypt_query", decryptDest)

	return nil
}

func encryptUserPII(c *security.AESCipher, u *models.User) error {
	var err error
	if u == nil {
		return nil
	}
	// Keep Email plaintext (uniqueness + lookup), Country plaintext (analytics/grouping),
	// FursonaName plaintext (display/search), and DateOfBirth plaintext (analytics queries).
	if u.FirstName, err = c.EncryptString(u.FirstName); err != nil {
		return err
	}
	if u.LastName, err = c.EncryptString(u.LastName); err != nil {
		return err
	}
	if u.IdCard, err = c.EncryptString(u.IdCard); err != nil {
		return err
	}
	return nil
}

func decryptUserPII(c *security.AESCipher, u *models.User) error {
	var err error
	if u == nil {
		return nil
	}
	if u.FirstName, err = c.DecryptString(u.FirstName); err != nil {
		return err
	}
	if u.LastName, err = c.DecryptString(u.LastName); err != nil {
		return err
	}
	if u.IdCard, err = c.DecryptString(u.IdCard); err != nil {
		return err
	}
	return nil
}

func walkAndApplyUsers(dest any, fn func(*models.User) error) error {
	if dest == nil {
		return nil
	}
	v := reflect.ValueOf(dest)
	return walkValue(v, fn)
}

func walkValue(v reflect.Value, fn func(*models.User) error) error {
	if !v.IsValid() {
		return nil
	}
	for v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		// Handle models.User or embedded.
		if u, ok := v.Addr().Interface().(*models.User); ok {
			return fn(u)
		}
		return nil

	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if err := walkValue(v.Index(i), fn); err != nil {
				return err
			}
		}
		return nil

	default:
		return nil
	}
}
