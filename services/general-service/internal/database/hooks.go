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
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}
		if rv.Kind() != reflect.Struct {
			return
		}

		idValue := field.ReflectValueOf(context.Background(), rv)
		if idValue.Kind() == reflect.Ptr {
			idValue = idValue.Elem()
		}

		// Nếu ID chưa có thì set UUID v7
		if idValue.IsValid() && idValue.CanSet() && idValue.Interface() == uuid.Nil {
			idValue.Set(reflect.ValueOf(uuid.Must(uuid.NewV7())))
		}
	})
}
