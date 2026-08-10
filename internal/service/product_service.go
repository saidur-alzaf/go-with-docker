package service

import (
	"context"
	"fmt"
	"io"

	"go-sqlite-api/internal/domain"
	"go-sqlite-api/internal/storage"
)

type ProductService interface {
	CreateProduct(ctx context.Context, product *domain.Product) error
	GetProductByID(ctx context.Context, id int64) (*domain.Product, error)
	GetAllProducts(ctx context.Context) ([]domain.Product, error)
	UpdateProduct(ctx context.Context, product *domain.Product) error
	DeleteProduct(ctx context.Context, id int64) error
	UploadProductImage(ctx context.Context, id int64, fileReader io.Reader, filename string, contentType string) (*domain.Product, error)
	CheckAndNotifyLowStock(ctx context.Context, product *domain.Product)
}

type productService struct {
	productRepo domain.ProductRepository
	notifier    domain.NotificationService
	storage     storage.StorageService
}

func NewProductService(productRepo domain.ProductRepository, notifier domain.NotificationService, storage storage.StorageService) ProductService {
	return &productService{
		productRepo: productRepo,
		notifier:    notifier,
		storage:     storage,
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

func (s *productService) UploadProductImage(ctx context.Context, id int64, fileReader io.Reader, filename string, contentType string) (*domain.Product, error) {
	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if s.storage == nil {
		return nil, fmt.Errorf("storage service is not configured")
	}

	imageURL, err := s.storage.UploadFile(ctx, fileReader, filename, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload image: %w", err)
	}

	product.ImageURL = imageURL
	if err := s.productRepo.Update(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to update product with image url: %w", err)
	}

	return product, nil
}

func (s *productService) CheckAndNotifyLowStock(ctx context.Context, product *domain.Product) {
	if product.StockQuantity <= 5 {
		go func() {
			msg := fmt.Sprintf("Warning: Low stock for '%s' (ID: %d). Only %d items remaining!", product.Name, product.ID, product.StockQuantity)
			_ = s.notifier.Send(context.Background(), "Low Stock Alert", msg, 8)
		}()
	}
}
