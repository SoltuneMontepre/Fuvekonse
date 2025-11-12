package database

import (
	"gorm.io/gorm"
)

var GlobalDB *gorm.DB

// CloseDB closes the global database connection if it exists.
func CloseDB() error {
	if GlobalDB != nil {
		db, err := GlobalDB.DB()
		if err != nil {
			return err
		}
		return db.Close()
	}
	return nil
}
