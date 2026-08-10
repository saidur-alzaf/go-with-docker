package config

import (
	"fmt"
	"os"
)

// DBConfig holds connection details for a single PostgreSQL database.
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// DSN returns the PostgreSQL connection string for this database.
func (d DBConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode)
}

type Config struct {
	Port        string
	UserDB      DBConfig
	ProductDB   DBConfig
	OrderDB     DBConfig
	GotifyURL   string
	GotifyToken string
	R2          R2Config
}

type R2Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	PublicURL       string
}

func LoadConfig() *Config {
	port := getEnvOrDefault("PORT", "8080")

	userDB := DBConfig{
		Host:     getEnvOrDefault("USER_DB_HOST", "localhost"),
		Port:     getEnvOrDefault("USER_DB_PORT", "5432"),
		User:     getEnvOrDefault("USER_DB_USER", "postgres"),
		Password: getEnvOrDefault("USER_DB_PASSWORD", "postgres"),
		DBName:   getEnvOrDefault("USER_DB_NAME", "user_db"),
		SSLMode:  getEnvOrDefault("USER_DB_SSLMODE", "disable"),
	}

	productDB := DBConfig{
		Host:     getEnvOrDefault("PRODUCT_DB_HOST", "localhost"),
		Port:     getEnvOrDefault("PRODUCT_DB_PORT", "5432"),
		User:     getEnvOrDefault("PRODUCT_DB_USER", "postgres"),
		Password: getEnvOrDefault("PRODUCT_DB_PASSWORD", "postgres"),
		DBName:   getEnvOrDefault("PRODUCT_DB_NAME", "product_db"),
		SSLMode:  getEnvOrDefault("PRODUCT_DB_SSLMODE", "disable"),
	}

	orderDB := DBConfig{
		Host:     getEnvOrDefault("ORDER_DB_HOST", "localhost"),
		Port:     getEnvOrDefault("ORDER_DB_PORT", "5432"),
		User:     getEnvOrDefault("ORDER_DB_USER", "postgres"),
		Password: getEnvOrDefault("ORDER_DB_PASSWORD", "postgres"),
		DBName:   getEnvOrDefault("ORDER_DB_NAME", "order_db"),
		SSLMode:  getEnvOrDefault("ORDER_DB_SSLMODE", "disable"),
	}

	r2Config := R2Config{
		AccountID:       os.Getenv("R2_ACCOUNT_ID"),
		AccessKeyID:     os.Getenv("R2_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("R2_SECRET_ACCESS_KEY"),
		BucketName:      os.Getenv("R2_BUCKET_NAME"),
		PublicURL:       os.Getenv("R2_PUBLIC_URL"),
	}

	return &Config{
		Port:        port,
		UserDB:      userDB,
		ProductDB:   productDB,
		OrderDB:     orderDB,
		GotifyURL:   os.Getenv("GOTIFY_URL"),
		GotifyToken: os.Getenv("GOTIFY_TOKEN"),
		R2:          r2Config,
	}
}

func getEnvOrDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
