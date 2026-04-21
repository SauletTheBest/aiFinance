package db

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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

    migrationsDir := "../../internal/db/migrations"

    log.Printf("Looking for migrations in: %s", migrationsDir)

	wd, _ := os.Getwd()     //to check your current dir
	log.Println("WORKDIR:", wd)

	log.Println("Checking migrations dir exists:", migrationsDir)

	if _, err := os.Stat(migrationsDir); err != nil {
		log.Println("MIGRATIONS DIR ERROR:", err)
	}

    files, err := filepath.Glob(filepath.Join(migrationsDir, "*.up.sql"))
    if err != nil {
        return err
    }

    sort.Strings(files)

    log.Printf("Found %d migration files: %v", len(files), files)

    for _, file := range files {
        log.Printf("Executing migration: %s", file)

        content, err := os.ReadFile(file) 
        if err != nil {
            return err
        }

        _, err = sqlDB.Exec(string(content))
        if err != nil {
            return err
        }

        log.Printf("Migration %s executed successfully", file)
    }

    return nil
}