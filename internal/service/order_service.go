package service

import (
	"context"
	"errors"
	"fmt"

	"go-sqlite-api/internal/domain"
)

type OrderService interface {
	CreateOrder(ctx context.Context, order *domain.Order) error
	GetOrderByID(ctx context.Context, id int64) (*domain.Order, error)
	GetAllOrders(ctx context.Context) ([]domain.Order, error)
	GetOrdersByUserID(ctx context.Context, userID int64) ([]domain.Order, error)
	UpdateOrderStatus(ctx context.Context, id int64, status string) error
	DeleteOrder(ctx context.Context, id int64) error
}

type orderService struct {
	orderRepo   domain.OrderRepository
	productRepo domain.ProductRepository
	userRepo    domain.UserRepository
	notifier    domain.NotificationService
}

func NewOrderService(
	orderRepo domain.OrderRepository,
	productRepo domain.ProductRepository,
	userRepo domain.UserRepository,
	notifier domain.NotificationService,
) OrderService {
	return &orderService{
		orderRepo:   orderRepo,
		productRepo: productRepo,
		userRepo:    userRepo,
		notifier:    notifier,
	}
}

func (s *orderService) CreateOrder(ctx context.Context, order *domain.Order) error {
	// Verify User exists
	user, err := s.userRepo.GetByID(ctx, order.UserID)
	if err != nil {
		return fmt.Errorf("user error: %w", err)
	}

	if len(order.Items) == 0 {
		return errors.New("order must contain at least one item")
	}

	var totalAmount float64

	// Validate product stock & calculate total
	for i := range order.Items {
		item := &order.Items[i]
		product, err := s.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			return fmt.Errorf("product ID %d not found: %w", item.ProductID, err)
		}

		if product.StockQuantity < item.Quantity {
			return fmt.Errorf("insufficient stock for product '%s'. requested: %d, available: %d", product.Name, item.Quantity, product.StockQuantity)
		}

		item.UnitPrice = product.Price
		totalAmount += product.Price * float64(item.Quantity)
	}

	order.TotalAmount = totalAmount

	// Deduct stock for each item
	for _, item := range order.Items {
		if err := s.productRepo.UpdateStock(ctx, item.ProductID, -item.Quantity); err != nil {
			return fmt.Errorf("failed to update product stock: %w", err)
		}
	}

	// Create Order in DB
	if err := s.orderRepo.Create(ctx, order); err != nil {
		return err
	}

	// Gotify Notification
	go func() {
		msg := fmt.Sprintf("New Order #%d placed by %s for $%.2f", order.ID, user.Name, order.TotalAmount)
		_ = s.notifier.Send(context.Background(), "New Order Placed", msg, 6)

		// Check if any product is now low on stock after deduction
		for _, item := range order.Items {
			p, err := s.productRepo.GetByID(context.Background(), item.ProductID)
			if err == nil && p.StockQuantity <= 5 {
				lowStockMsg := fmt.Sprintf("Low stock alert for '%s' (ID: %d): %d items left after Order #%d", p.Name, p.ID, p.StockQuantity, order.ID)
				_ = s.notifier.Send(context.Background(), "Low Stock Alert", lowStockMsg, 8)
			}
		}
	}()

	return nil
}

func (s *orderService) GetOrderByID(ctx context.Context, id int64) (*domain.Order, error) {
	return s.orderRepo.GetByID(ctx, id)
}

func (s *orderService) GetAllOrders(ctx context.Context) ([]domain.Order, error) {
	return s.orderRepo.GetAll(ctx)
}

func (s *orderService) GetOrdersByUserID(ctx context.Context, userID int64) ([]domain.Order, error) {
	return s.orderRepo.GetByUserID(ctx, userID)
}

func (s *orderService) UpdateOrderStatus(ctx context.Context, id int64, status string) error {
	if err := s.orderRepo.UpdateStatus(ctx, id, status); err != nil {
		return err
	}

	go func() {
		msg := fmt.Sprintf("Order #%d status updated to '%s'", id, status)
		_ = s.notifier.Send(context.Background(), "Order Status Update", msg, 5)
	}()

	return nil
}

func (s *orderService) DeleteOrder(ctx context.Context, id int64) error {
	return s.orderRepo.Delete(ctx, id)
}
