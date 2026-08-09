package handler

import (
	"net/http"
	"strconv"

	"go-sqlite-api/internal/domain"
	"go-sqlite-api/internal/service"
)

type AnalysisHandler struct {
	analysisService service.AnalysisService
}

func NewAnalysisHandler(analysisService service.AnalysisService) *AnalysisHandler {
	return &AnalysisHandler{analysisService: analysisService}
}

func (h *AnalysisHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	stats, err := h.analysisService.GetSummaryStats(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

func (h *AnalysisHandler) GetTopProducts(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 5
	}

	products, err := h.analysisService.GetTopProducts(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if products == nil {
		products = []domain.TopProduct{}
	}
	writeJSON(w, http.StatusOK, products)
}

func (h *AnalysisHandler) GetUserSpending(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 5
	}

	spendings, err := h.analysisService.GetUserSpending(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if spendings == nil {
		spendings = []domain.UserSpending{}
	}
	writeJSON(w, http.StatusOK, spendings)
}

func (h *AnalysisHandler) GetSalesByStatus(w http.ResponseWriter, r *http.Request) {
	trends, err := h.analysisService.GetSalesByStatus(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if trends == nil {
		trends = []domain.SalesTrend{}
	}
	writeJSON(w, http.StatusOK, trends)
}
