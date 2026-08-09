package domain

import (
	"context"
)

type SummaryStats struct {
	TotalUsers      int64   `json:"total_users"`
	TotalProducts   int64   `json:"total_products"`
	TotalOrders     int64   `json:"total_orders"`
	TotalRevenue    float64 `json:"total_revenue"`
	AverageOrderVal float64 `json:"average_order_value"`
	LowStockCount   int64   `json:"low_stock_products_count"`
}

type TopProduct struct {
	ProductID   int64   `json:"product_id"`
	ProductName string  `json:"product_name"`
	TotalSold   int64   `json:"total_sold"`
	TotalRevenue float64 `json:"total_revenue"`
}

type UserSpending struct {
	UserID     int64   `json:"user_id"`
	UserName   string  `json:"user_name"`
	UserEmail  string  `json:"user_email"`
	OrderCount int64   `json:"order_count"`
	TotalSpent float64 `json:"total_spent"`
}

type SalesTrend struct {
	Status      string  `json:"status"`
	OrderCount  int64   `json:"order_count"`
	TotalAmount float64 `json:"total_amount"`
}

type AnalysisRepository interface {
	GetSummaryStats(ctx context.Context) (*SummaryStats, error)
	GetTopProducts(ctx context.Context, limit int) ([]TopProduct, error)
	GetUserSpending(ctx context.Context, limit int) ([]UserSpending, error)
	GetSalesByStatus(ctx context.Context) ([]SalesTrend, error)
}
