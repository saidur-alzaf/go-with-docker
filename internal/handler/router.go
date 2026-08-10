package handler

import (
	"log"
	"net/http"
	"time"
)

type RouterDependencies struct {
	UserHandler     *UserHandler
	ProductHandler  *ProductHandler
	OrderHandler    *OrderHandler
	AnalysisHandler *AnalysisHandler
}

func SetupRouter(deps RouterDependencies) http.Handler {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// User Routes
	mux.HandleFunc("POST /api/v1/users", deps.UserHandler.Create)
	mux.HandleFunc("GET /api/v1/users", deps.UserHandler.GetAll)
	mux.HandleFunc("GET /api/v1/users/{id}", deps.UserHandler.GetByID)
	mux.HandleFunc("PUT /api/v1/users/{id}", deps.UserHandler.Update)
	mux.HandleFunc("DELETE /api/v1/users/{id}", deps.UserHandler.Delete)

	// Product Routes
	mux.HandleFunc("POST /api/v1/products", deps.ProductHandler.Create)
	mux.HandleFunc("GET /api/v1/products", deps.ProductHandler.GetAll)
	mux.HandleFunc("GET /api/v1/products/{id}", deps.ProductHandler.GetByID)
	mux.HandleFunc("PUT /api/v1/products/{id}", deps.ProductHandler.Update)
	mux.HandleFunc("DELETE /api/v1/products/{id}", deps.ProductHandler.Delete)
	mux.HandleFunc("POST /api/v1/products/{id}/image", deps.ProductHandler.UploadImage)

	// Order Routes
	mux.HandleFunc("POST /api/v1/orders", deps.OrderHandler.Create)
	mux.HandleFunc("GET /api/v1/orders", deps.OrderHandler.GetAll)
	mux.HandleFunc("GET /api/v1/orders/{id}", deps.OrderHandler.GetByID)
	mux.HandleFunc("PATCH /api/v1/orders/{id}/status", deps.OrderHandler.UpdateStatus)
	mux.HandleFunc("DELETE /api/v1/orders/{id}", deps.OrderHandler.Delete)

	// Analysis / Analytics Routes
	mux.HandleFunc("GET /api/v1/analysis/summary", deps.AnalysisHandler.GetSummary)
	mux.HandleFunc("GET /api/v1/analysis/top-products", deps.AnalysisHandler.GetTopProducts)
	mux.HandleFunc("GET /api/v1/analysis/user-stats", deps.AnalysisHandler.GetUserSpending)
	mux.HandleFunc("GET /api/v1/analysis/sales-trend", deps.AnalysisHandler.GetSalesByStatus)

	return loggingMiddleware(mux)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("started %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
		log.Printf("completed %s %s in %v", r.Method, r.URL.Path, time.Since(start))
	})
}
