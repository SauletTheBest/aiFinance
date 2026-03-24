package db


import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"path/filepath"
	"io/fs"
	"fmt"
	"log"
	"os"
)

func NewPostgres(dsn string) (*gorm.DB, error) {
	
	db,  err :=  gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func RunMigrations(db *gorm.DB) error {
	// Get the database connection
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	// Read all migration files
	migrationsDir := "internal/db/migrations"
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.up.sql"))
	if err != nil {
		return fmt.Errorf("failed to read migration files: %w", err)
	}

	// Execute migrations in order
	for _, file := range files {
		content, err := fs.ReadFile(os.DirFS("."), file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		log.Printf("Executing migration: %s", file)
		_, err = sqlDB.Exec(string(content))

		if err !=  nil {
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}
	}

	log.Println("All migrations completed successfully")
	return nil
}