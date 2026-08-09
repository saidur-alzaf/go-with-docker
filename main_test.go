package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go-sqlite-api/internal/handler"
)

func TestHealthEndpoint(t *testing.T) {
	router := handler.SetupRouter(handler.RouterDependencies{})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}
