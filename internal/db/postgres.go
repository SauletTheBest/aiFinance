package db

import (
	"embed"
	"log"
	"sort"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//go:embed migrations/*.up.sql
var migrationFiles embed.FS

func NewPostgres(dsn string) (*gorm.DB, error) {

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func RunMigrations(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	// Read files from our embedded filesystem (no disk paths needed!)
	files, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return err
	}

	var migrationFileNames []string
	for _, file := range files {
		// Only run '.up.sql' files
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".up.sql") {
			migrationFileNames = append(migrationFileNames, file.Name())
		}
	}

	// Sort them alphabetically to ensure they execute in order (000001, 000002, etc.)
	sort.Strings(migrationFileNames)

	log.Printf("Found %d embedded up-migration files", len(migrationFileNames))

	for _, fileName := range migrationFileNames {
		log.Printf("Executing embedded migration: %s", fileName)

		content, err := migrationFiles.ReadFile("migrations/" + fileName)
		if err != nil {
			return err
		}

		_, err = sqlDB.Exec(string(content))
		if err != nil {
			return err
		}

		log.Printf("Migration %s executed successfully", fileName)
	}

	return nil
}