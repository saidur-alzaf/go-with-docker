package service

import (
	"context"
	"fmt"

	"go-sqlite-api/internal/domain"
)

type ProductService interface {
	CreateProduct(ctx context.Context, product *domain.Product) error
	GetProductByID(ctx context.Context, id int64) (*domain.Product, error)
	GetAllProducts(ctx context.Context) ([]domain.Product, error)
	UpdateProduct(ctx context.Context, product *domain.Product) error
	DeleteProduct(ctx context.Context, id int64) error
	CheckAndNotifyLowStock(ctx context.Context, product *domain.Product)
}

type productService struct {
	productRepo domain.ProductRepository
	notifier    domain.NotificationService
}

func NewProductService(productRepo domain.ProductRepository, notifier domain.NotificationService) ProductService {
	return &productService{
		productRepo: productRepo,
		notifier:    notifier,
	}
}

func (s *productService) CreateProduct(ctx context.Context, product *domain.Product) error {
	if err := s.productRepo.Create(ctx, product); err != nil {
		return err
	}
	s.CheckAndNotifyLowStock(ctx, product)
	return nil
}

func (s *productService) GetProductByID(ctx context.Context, id int64) (*domain.Product, error) {
	return s.productRepo.GetByID(ctx, id)
}

func (s *productService) GetAllProducts(ctx context.Context) ([]domain.Product, error) {
	return s.productRepo.GetAll(ctx)
}

func (s *productService) UpdateProduct(ctx context.Context, product *domain.Product) error {
	if err := s.productRepo.Update(ctx, product); err != nil {
		return err
	}
	s.CheckAndNotifyLowStock(ctx, product)
	return nil
}

func (s *productService) DeleteProduct(ctx context.Context, id int64) error {
	return s.productRepo.Delete(ctx, id)
}

func (s *productService) CheckAndNotifyLowStock(ctx context.Context, product *domain.Product) {
	if product.StockQuantity <= 5 {
		go func() {
			msg := fmt.Sprintf("Warning: Low stock for '%s' (ID: %d). Only %d items remaining!", product.Name, product.ID, product.StockQuantity)
			_ = s.notifier.Send(context.Background(), "Low Stock Alert", msg, 8)
		}()
	}
}
