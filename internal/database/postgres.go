package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

// ConnectDB opens and pings a PostgreSQL connection using the given DSN.
// The label is used for log messages to identify which database this is.
func ConnectDB(dsn, label string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s database: %w", label, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping %s database: %w", label, err)
	}

	log.Printf("Successfully connected to %s PostgreSQL database", label)
	return db, nil
}

// MigrateUserDB creates the users table in the user database.
func MigrateUserDB(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		email VARCHAR(255) UNIQUE NOT NULL,
		role VARCHAR(50) NOT NULL DEFAULT 'customer',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("failed to migrate user schema: %w", err)
	}
	log.Println("User DB migration completed successfully")
	return nil
}

// MigrateProductDB creates the products table in the product database.
func MigrateProductDB(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS products (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		price NUMERIC(10, 2) NOT NULL DEFAULT 0.00,
		stock_quantity INT NOT NULL DEFAULT 0,
		category VARCHAR(100) DEFAULT 'general',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("failed to migrate product schema: %w", err)
	}
	log.Println("Product DB migration completed successfully")
	return nil
}

// MigrateOrderDB creates the orders and order_items tables in the order database.
// No cross-database foreign keys — referential integrity is enforced at the service layer.
func MigrateOrderDB(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS orders (
		id SERIAL PRIMARY KEY,
		user_id INT NOT NULL,
		total_amount NUMERIC(10, 2) NOT NULL DEFAULT 0.00,
		status VARCHAR(50) NOT NULL DEFAULT 'pending',
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS order_items (
		id SERIAL PRIMARY KEY,
		order_id INT NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
		product_id INT NOT NULL,
		quantity INT NOT NULL DEFAULT 1,
		unit_price NUMERIC(10, 2) NOT NULL
	);
	`
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("failed to migrate order schema: %w", err)
	}
	log.Println("Order DB migration completed successfully")
	return nil
}
