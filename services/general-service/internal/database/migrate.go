package database

import (
	"fmt"
	"general-service/internal/models"
	"log"

	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	// List of models to migrate
	allModels := []interface{}{
		&models.User{},
		&models.DealerBooth{},
		&models.UserDealerStaff{},
		&models.TicketTier{},
		&models.UserTicket{},
		&models.ConBookArt{},
		&models.Payment{},
	}

	// AutoMigrate (creates tables, adds columns, indexes)
	err := db.AutoMigrate(allModels...)
	if err != nil {
		return fmt.Errorf("failed to auto-migrate base tables: %w", err)
	}

	// Ensure google_id exists on users (handles DBs created before this column was added, e.g. CI)
	if err := ensureUsersGoogleIdColumn(db); err != nil {
		return fmt.Errorf("failed to ensure users.google_id column: %w", err)
	}

	// Drop columns that are no longer in the model
	// WARNING: This will permanently DELETE DATA!
	if err := dropUnusedColumns(db, allModels); err != nil {
		return fmt.Errorf("failed to drop unused columns: %w", err)
	}

	err = db.AutoMigrate()
	if err != nil {
		return fmt.Errorf("failed to auto-migrate dependent tables: %w", err)
	}

	log.Println("Database migration completed successfully")
	return nil
}

// ensureUsersGoogleIdColumn adds the google_id column to users if it does not exist.
// This covers databases created before GoogleId was added to the User model (e.g. CI).
func ensureUsersGoogleIdColumn(db *gorm.DB) error {
	migrator := db.Migrator()
	if migrator.HasTable("users") && !migrator.HasColumn("users", "google_id") {
		if err := migrator.AddColumn(&models.User{}, "GoogleId"); err != nil {
			return err
		}
		log.Println("Added users.google_id column (migration)")
	}
	return nil
}

// dropUnusedColumns removes columns that are no longer in the model
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
	// Get table name
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

	// Get all current columns in the database
	columnTypes, err := migrator.ColumnTypes(tableName)
	if err != nil {
		return fmt.Errorf("failed to get column types for table %s: %w", tableName, err)
	}

	// Get list of fields in the model
	modelFields := make(map[string]bool, len(stmt.Schema.Fields))
	for _, field := range stmt.Schema.Fields {
		modelFields[field.DBName] = true
	}

	// Drop columns that are no longer in the model
	for _, columnType := range columnTypes {
		columnName := columnType.Name()
		if !modelFields[columnName] {
			log.Printf("Dropping column %s from table %s", columnName, tableName)
			if err := migrator.DropColumn(model, columnName); err != nil {
				return fmt.Errorf("failed to drop column %s from %s: %w", columnName, tableName, err)
			}
		}
	}

	return nil
}

// migrateTicketTierCodes assigns tier_code field to tiers that don't have one yet.
// Once a tier_code is set, it becomes immutable to ensure ticket reference codes remain stable.
func migrateTicketTierCodes(db *gorm.DB) error {
	// Check if ticket_tiers table exists
	if !db.Migrator().HasTable("ticket_tiers") {
		log.Println("ticket_tiers table doesn't exist yet, skipping tier code migration")
		return nil
	}

	// Only assign tier codes to rows where tier_code IS NULL.
	// This ensures existing tier codes are never overwritten, keeping ticket references stable.
	// New tiers get the next available T{n} code based on current max.
	sql := `
		WITH null_tiers AS (
			SELECT id, ROW_NUMBER() OVER (ORDER BY price ASC) as row_num
			FROM ticket_tiers
			WHERE is_deleted = false AND tier_code IS NULL
		),
		max_existing AS (
			SELECT COALESCE(MAX(CAST(SUBSTRING(tier_code FROM 2) AS INTEGER)), 0) as max_num
			FROM ticket_tiers
			WHERE tier_code IS NOT NULL AND tier_code ~ '^T[0-9]+$'
		)
		UPDATE ticket_tiers t
		SET tier_code = 'T' || (nt.row_num + me.max_num)
		FROM null_tiers nt, max_existing me
		WHERE t.id = nt.id;
	`

	result := db.Exec(sql)
	if result.Error != nil {
		return fmt.Errorf("failed to update tier codes: %w", result.Error)
	}

	if result.RowsAffected > 0 {
		log.Printf("Assigned tier codes to %d new ticket tiers", result.RowsAffected)
	} else {
		log.Println("No new ticket tiers need tier codes")
	}

	return nil
}

func MigrateAndSeed(db *gorm.DB) error {
	if err := AutoMigrate(db); err != nil {
		return err
	}

	if err := migrateTicketTierCodes(db); err != nil {
		return fmt.Errorf("failed to migrate ticket tier codes: %w", err)
	}

	if err := seedInitialData(); err != nil {
		return fmt.Errorf("failed to seed data: %w", err)
	}

	return nil
}

func seedInitialData() error {
	// TODO: Implement initial data seeding logic
	// Currently, no initial data seeding is required
	log.Println("Initial data seeding skipped (no data configured)")
	return nil
}
