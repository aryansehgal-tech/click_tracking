package db

import (
	"click_tracking/internal/models"
	"fmt"
	"log"
	"os"
	"time"

	// "github.com/Ryana-X-x/click_tracking/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgres() (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
	os.Getenv("DB_HOST"),
	os.Getenv("DB_USER"),
	os.Getenv("DB_PASSWORD"),
	os.Getenv("DB_NAME"),
	os.Getenv("DB_PORT"),
)
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: true,
	})

	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Connection Pool Settings
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(30 * time.Minute)

	// Auto Migrate Models
	if err := db.AutoMigrate(
		&models.Session{},
		&models.Event{},
	); err != nil {
		return nil, err
	}

	log.Println("PostgreSQL connected and migrated successfully")

	return db, nil
}
