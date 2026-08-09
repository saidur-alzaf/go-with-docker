package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"go-sqlite-api/internal/domain"
)

type productRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) domain.ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(ctx context.Context, product *domain.Product) error {
	query := `
		INSERT INTO products (name, description, price, stock_quantity, category, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	now := time.Now()
	product.CreatedAt = now
	product.UpdatedAt = now

	err := r.db.QueryRowContext(ctx, query,
		product.Name, product.Description, product.Price, product.StockQuantity, product.Category, product.CreatedAt, product.UpdatedAt,
	).Scan(&product.ID)
	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}
	return nil
}

func (r *productRepository) GetByID(ctx context.Context, id int64) (*domain.Product, error) {
	query := `SELECT id, name, description, price, stock_quantity, category, created_at, updated_at FROM products WHERE id = $1`
	var product domain.Product
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&product.ID, &product.Name, &product.Description, &product.Price, &product.StockQuantity, &product.Category, &product.CreatedAt, &product.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("product not found")
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	return &product, nil
}

func (r *productRepository) GetAll(ctx context.Context) ([]domain.Product, error) {
	query := `SELECT id, name, description, price, stock_quantity, category, created_at, updated_at FROM products ORDER BY id ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var product domain.Product
		if err := rows.Scan(
			&product.ID, &product.Name, &product.Description, &product.Price, &product.StockQuantity, &product.Category, &product.CreatedAt, &product.UpdatedAt,
		); err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, nil
}

func (r *productRepository) Update(ctx context.Context, product *domain.Product) error {
	query := `
		UPDATE products
		SET name = $1, description = $2, price = $3, stock_quantity = $4, category = $5, updated_at = $6
		WHERE id = $7
	`
	product.UpdatedAt = time.Now()
	res, err := r.db.ExecContext(ctx, query,
		product.Name, product.Description, product.Price, product.StockQuantity, product.Category, product.UpdatedAt, product.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("product not found")
	}
	return nil
}

func (r *productRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM products WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("product not found")
	}
	return nil
}

func (r *productRepository) UpdateStock(ctx context.Context, id int64, stockDelta int) error {
	query := `
		UPDATE products
		SET stock_quantity = stock_quantity + $1, updated_at = $2
		WHERE id = $3 AND (stock_quantity + $1) >= 0
	`
	res, err := r.db.ExecContext(ctx, query, stockDelta, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update product stock: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("insufficient stock or product not found")
	}
	return nil
}
