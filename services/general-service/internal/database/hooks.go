package database

import (
	"context"
	"reflect"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func AutoGenerateUUID(db *gorm.DB) {
	db.Callback().Create().Before("gorm:create").Register("auto_uuid_v7", func(tx *gorm.DB) {
		if tx.Statement == nil || tx.Statement.Schema == nil {
			return
		}

		field := tx.Statement.Schema.LookUpField("Id")
		if field == nil {
			return
		}

		rv := tx.Statement.ReflectValue
		if rv.Kind() == reflect.Pointer {
			rv = rv.Elem()
		}
		if rv.Kind() != reflect.Struct {
			return
		}

		idValue := field.ReflectValueOf(context.Background(), rv)
		if idValue.Kind() == reflect.Pointer {
			idValue = idValue.Elem()
		}

		// Set UUID v7 if ID is not yet assigned
		if idValue.IsValid() && idValue.CanSet() && idValue.Interface() == uuid.Nil {
			newUUID, err := uuid.NewV7()
			if err != nil {
				return // or log error
			}
			idValue.Set(reflect.ValueOf(newUUID))
		}
	})
}
