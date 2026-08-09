package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"go-sqlite-api/internal/domain"
)

// analysisRepository queries across all 3 databases and aggregates results in Go.
type analysisRepository struct {
	userDB    *sql.DB
	productDB *sql.DB
	orderDB   *sql.DB
}

func NewAnalysisRepository(userDB, productDB, orderDB *sql.DB) domain.AnalysisRepository {
	return &analysisRepository{
		userDB:    userDB,
		productDB: productDB,
		orderDB:   orderDB,
	}
}

func (r *analysisRepository) GetSummaryStats(ctx context.Context) (*domain.SummaryStats, error) {
	var stats domain.SummaryStats

	// Total users — from userDB
	if err := r.userDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&stats.TotalUsers); err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	// Total products — from productDB
	if err := r.productDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM products`).Scan(&stats.TotalProducts); err != nil {
		return nil, fmt.Errorf("failed to count products: %w", err)
	}

	// Total orders — from orderDB
	if err := r.orderDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM orders`).Scan(&stats.TotalOrders); err != nil {
		return nil, fmt.Errorf("failed to count orders: %w", err)
	}

	// Total revenue & average order value — from orderDB
	if err := r.orderDB.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(total_amount), 0), COALESCE(AVG(total_amount), 0)
		FROM orders
		WHERE status != 'cancelled'
	`).Scan(&stats.TotalRevenue, &stats.AverageOrderVal); err != nil {
		return nil, fmt.Errorf("failed to calculate revenue stats: %w", err)
	}

	// Low stock products count — from productDB
	if err := r.productDB.QueryRowContext(ctx, `SELECT COUNT(*) FROM products WHERE stock_quantity <= 5`).Scan(&stats.LowStockCount); err != nil {
		return nil, fmt.Errorf("failed to count low stock products: %w", err)
	}

	return &stats, nil
}

func (r *analysisRepository) GetTopProducts(ctx context.Context, limit int) ([]domain.TopProduct, error) {
	if limit <= 0 {
		limit = 5
	}

	// Step 1: Aggregate sales data from orderDB (order_items + orders)
	query := `
		SELECT oi.product_id,
		       COALESCE(SUM(oi.quantity), 0) AS total_sold,
		       COALESCE(SUM(oi.quantity * oi.unit_price), 0) AS total_revenue
		FROM order_items oi
		JOIN orders o ON oi.order_id = o.id
		WHERE o.status != 'cancelled'
		GROUP BY oi.product_id
		ORDER BY total_sold DESC
		LIMIT $1
	`
	rows, err := r.orderDB.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top products from order DB: %w", err)
	}
	defer rows.Close()

	type salesRow struct {
		ProductID    int64
		TotalSold    int64
		TotalRevenue float64
	}
	var salesData []salesRow
	for rows.Next() {
		var s salesRow
		if err := rows.Scan(&s.ProductID, &s.TotalSold, &s.TotalRevenue); err != nil {
			return nil, err
		}
		salesData = append(salesData, s)
	}

	// Step 2: Look up product names from productDB
	var products []domain.TopProduct
	for _, s := range salesData {
		var name string
		err := r.productDB.QueryRowContext(ctx, `SELECT name FROM products WHERE id = $1`, s.ProductID).Scan(&name)
		if err != nil {
			name = fmt.Sprintf("Product #%d (not found)", s.ProductID)
		}
		products = append(products, domain.TopProduct{
			ProductID:    s.ProductID,
			ProductName:  name,
			TotalSold:    s.TotalSold,
			TotalRevenue: s.TotalRevenue,
		})
	}

	return products, nil
}

func (r *analysisRepository) GetUserSpending(ctx context.Context, limit int) ([]domain.UserSpending, error) {
	if limit <= 0 {
		limit = 5
	}

	// Step 1: Aggregate spending data from orderDB
	query := `
		SELECT user_id, COUNT(id) AS order_count, COALESCE(SUM(total_amount), 0) AS total_spent
		FROM orders
		WHERE status != 'cancelled'
		GROUP BY user_id
		ORDER BY total_spent DESC
		LIMIT $1
	`
	rows, err := r.orderDB.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get user spending from order DB: %w", err)
	}
	defer rows.Close()

	type spendingRow struct {
		UserID     int64
		OrderCount int64
		TotalSpent float64
	}
	var spendingData []spendingRow
	for rows.Next() {
		var s spendingRow
		if err := rows.Scan(&s.UserID, &s.OrderCount, &s.TotalSpent); err != nil {
			return nil, err
		}
		spendingData = append(spendingData, s)
	}

	// Step 2: Look up user details from userDB
	var spendings []domain.UserSpending
	for _, s := range spendingData {
		var name, email string
		err := r.userDB.QueryRowContext(ctx, `SELECT name, email FROM users WHERE id = $1`, s.UserID).Scan(&name, &email)
		if err != nil {
			name = fmt.Sprintf("User #%d (not found)", s.UserID)
			email = "unknown"
		}
		spendings = append(spendings, domain.UserSpending{
			UserID:     s.UserID,
			UserName:   name,
			UserEmail:  email,
			OrderCount: s.OrderCount,
			TotalSpent: s.TotalSpent,
		})
	}

	return spendings, nil
}

func (r *analysisRepository) GetSalesByStatus(ctx context.Context) ([]domain.SalesTrend, error) {
	// Only needs orderDB — no cross-DB query
	query := `
		SELECT status, COUNT(id) AS order_count, COALESCE(SUM(total_amount), 0) AS total_amount
		FROM orders
		GROUP BY status
		ORDER BY order_count DESC
	`
	rows, err := r.orderDB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get sales by status: %w", err)
	}
	defer rows.Close()

	var trends []domain.SalesTrend
	for rows.Next() {
		var t domain.SalesTrend
		if err := rows.Scan(&t.Status, &t.OrderCount, &t.TotalAmount); err != nil {
			return nil, err
		}
		trends = append(trends, t)
	}
	return trends, nil
}
