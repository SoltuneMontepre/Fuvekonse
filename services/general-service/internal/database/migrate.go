package database

import (
	"fmt"
	"general-service/internal/models"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	// Danh sách các models cần migrate
	allModels := []interface{}{
		&models.User{},
		&models.DealerBooth{},
		&models.UserDealerStaff{},
		&models.TicketTier{},
		&models.UserTicket{},
		&models.ConBookArt{},
		&models.Payment{},
	}

	// AutoMigrate thông thường (tạo tables, thêm columns, indexes)
	err := db.AutoMigrate(allModels...)
	if err != nil {
		return fmt.Errorf("failed to auto-migrate base tables: %w", err)
	}

	// Drop các columns không còn trong model
	// CẢNH BÁO: Điều này sẽ XÓA DỮ LIỆU vĩnh viễn!
	if err := dropUnusedColumns(db, allModels); err != nil {
		return fmt.Errorf("failed to drop unused columns: %w", err)
	}

	err = db.AutoMigrate()
	if err != nil {
		return fmt.Errorf("failed to auto-migrate dependent tables: %w", err)
	}

	fmt.Println("Database migration completed successfully")
	return nil
}

// dropUnusedColumns xóa các columns không còn trong model
func dropUnusedColumns(db *gorm.DB, models []interface{}) error {
	migrator := db.Migrator()

	for _, model := range models {
		if err := dropColumnsForModel(db, migrator, model); err != nil {
			return err
		}
	}

	return nil
}

// dropColumnsForModel handles dropping unused columns for a single model.
func dropColumnsForModel(db *gorm.DB, migrator gorm.Migrator, model interface{}) error {
	// Lấy tên table
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(model); err != nil {
		return err
	}
	tableName := stmt.Schema.Table

	// Check if table exists before attempting to get column types
	if !migrator.HasTable(tableName) {
		// Table doesn't exist yet, skip
		return nil
	}

	// Lấy tất cả columns hiện tại trong database
	columnTypes, err := migrator.ColumnTypes(tableName)
	if err != nil {
		return fmt.Errorf("failed to get column types for table %s: %w", tableName, err)
	}

	// Lấy danh sách fields trong model
	modelFields := make(map[string]bool, len(stmt.Schema.Fields))
	for _, field := range stmt.Schema.Fields {
		modelFields[field.DBName] = true
	}

	// Drop các columns không còn trong model
	for _, columnType := range columnTypes {
		columnName := columnType.Name()
		if !modelFields[columnName] {
			fmt.Printf("Dropping column %s from table %s\n", columnName, tableName)
			if err := migrator.DropColumn(model, columnName); err != nil {
				return fmt.Errorf("failed to drop column %s from %s: %w", columnName, tableName, err)
			}
		}
	}

	return nil
}

func MigrateAndSeed(db *gorm.DB) error {
	if err := AutoMigrate(db); err != nil {
		return err
	}

	if err := seedInitialData(); err != nil {
		return fmt.Errorf("failed to seed data: %w", err)
	}

	return nil
}

func seedInitialData() error {

	fmt.Println("Initial data seeded successfully")
	return nil
}
