package config

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"event-api/models"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	ServerPort string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "eventdb"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
	}
}

// getEnv gets environment variable or returns default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// ConnectDB establishes database connection using GORM
func (c *Config) ConnectDB() (*gorm.DB, error) {
	// First, connect to postgres database to create our database if it doesn't exist
	defaultDSN := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword,
	)

	defaultDB, err := gorm.Open(postgres.Open(defaultDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	// Create database if not exists
	var count int64
	defaultDB.Raw("SELECT COUNT(*) FROM pg_database WHERE datname = ?", c.DBName).Scan(&count)
	if count == 0 {
		log.Printf("Creating database: %s", c.DBName)
		defaultDB.Exec(fmt.Sprintf("CREATE DATABASE %s", c.DBName))
	}

	// Close default connection
	sqlDB, _ := defaultDB.DB()
	sqlDB.Close()

	// Now connect to our specific database
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate the schema
	if err := db.AutoMigrate(
		&models.User{},
		&models.Event{},
		&models.Registration{},
	); err != nil {
		return nil, fmt.Errorf("failed to auto migrate: %w", err)
	}

	log.Println("Database connection established and migrations completed")
	return db, nil
}

// GetDSN returns the Data Source Name for external use
func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName,
	)
}
