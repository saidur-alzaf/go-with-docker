package service

import (
	"context"

	"go-sqlite-api/internal/domain"
)

type AnalysisService interface {
	GetSummaryStats(ctx context.Context) (*domain.SummaryStats, error)
	GetTopProducts(ctx context.Context, limit int) ([]domain.TopProduct, error)
	GetUserSpending(ctx context.Context, limit int) ([]domain.UserSpending, error)
	GetSalesByStatus(ctx context.Context) ([]domain.SalesTrend, error)
}

type analysisService struct {
	analysisRepo domain.AnalysisRepository
}

func NewAnalysisService(analysisRepo domain.AnalysisRepository) AnalysisService {
	return &analysisService{
		analysisRepo: analysisRepo,
	}
}

func (s *analysisService) GetSummaryStats(ctx context.Context) (*domain.SummaryStats, error) {
	return s.analysisRepo.GetSummaryStats(ctx)
}

func (s *analysisService) GetTopProducts(ctx context.Context, limit int) ([]domain.TopProduct, error) {
	return s.analysisRepo.GetTopProducts(ctx, limit)
}

func (s *analysisService) GetUserSpending(ctx context.Context, limit int) ([]domain.UserSpending, error) {
	return s.analysisRepo.GetUserSpending(ctx, limit)
}

func (s *analysisService) GetSalesByStatus(ctx context.Context) ([]domain.SalesTrend, error) {
	return s.analysisRepo.GetSalesByStatus(ctx)
}
