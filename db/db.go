package db

import (
	"embed"
	"fmt"
	"sort"
	"strings"

	"github.com/ziaulkamal/zycrypt/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

var DB *gorm.DB

func Connect() error {
	var err error
	DB, err = gorm.Open(postgres.Open(config.C.Database.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)

	return nil
}

func Migrate() error {
	if DB == nil {
		return fmt.Errorf("database not connected")
	}

	// Ensure migrations table exists
	if err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(50) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT NOW()
		)
	`).Error; err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, file := range files {
		version := strings.TrimSuffix(file, ".sql")

		var count int64
		DB.Raw("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version).Scan(&count)
		if count > 0 {
			continue
		}

		content, err := migrationsFS.ReadFile("migrations/" + file)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", file, err)
		}

		if err := DB.Exec(string(content)).Error; err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", file, err)
		}

		DB.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version)
		fmt.Printf("  ✓ Applied: %s\n", file)
	}

	return nil
}
