// Package config/config.go
package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Config represents the application configuration
type Config struct {
	DB *gorm.DB
}

// Load loads the configuration from environment variables or .env file
func Load() *Config {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	// Get database configuration from environment variables
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// Create database connection string
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName)

	// Open database connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the database")
	}

	// Auto-migrate models
	// Add your models here
	// Example: db.AutoMigrate(&models.User{}, &models.Organization{}, ...)

	// Return the configuration
	return &Config{
		DB: db,
	}
}
