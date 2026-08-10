package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"go-sqlite-api/internal/config"
	"go-sqlite-api/internal/database"
	"go-sqlite-api/internal/handler"
	"go-sqlite-api/internal/notification"
	"go-sqlite-api/internal/repository/postgres"
	"go-sqlite-api/internal/service"
	"go-sqlite-api/internal/storage"
)

func main() {
	logger := log.New(os.Stdout, "[APP] ", log.LstdFlags|log.Lshortfile)

	cfg := config.LoadConfig()

	userDB, err := database.ConnectDB(cfg.UserDB.DSN(), "User")
	if err != nil {
		logger.Fatalf("User DB initialization error: %v", err)
	}
	defer userDB.Close()

	productDB, err := database.ConnectDB(cfg.ProductDB.DSN(), "Product")
	if err != nil {
		logger.Fatalf("Product DB initialization error: %v", err)
	}
	defer productDB.Close()

	orderDB, err := database.ConnectDB(cfg.OrderDB.DSN(), "Order")
	if err != nil {
		logger.Fatalf("Order DB initialization error: %v", err)
	}
	defer orderDB.Close()

	if err := database.MigrateUserDB(userDB); err != nil {
		logger.Fatalf("User DB migration error: %v", err)
	}
	if err := database.MigrateProductDB(productDB); err != nil {
		logger.Fatalf("Product DB migration error: %v", err)
	}
	if err := database.MigrateOrderDB(orderDB); err != nil {
		logger.Fatalf("Order DB migration error: %v", err)
	}

	userRepo := postgres.NewUserRepository(userDB)
	productRepo := postgres.NewProductRepository(productDB)
	orderRepo := postgres.NewOrderRepository(orderDB)
	analysisRepo := postgres.NewAnalysisRepository(userDB, productDB, orderDB)

	gotifyService := notification.NewGotifyService(cfg.GotifyURL, cfg.GotifyToken)

	var storageService storage.StorageService
	if cfg.R2.AccountID != "" {
		r2Storage, err := storage.NewR2StorageService(context.Background(), cfg.R2)
		if err != nil {
			logger.Printf("Warning: Failed to initialize R2 storage: %v", err)
		} else {
			storageService = r2Storage
			logger.Println("Cloudflare R2 storage service initialized successfully")
		}
	} else {
		logger.Println("R2 storage configuration missing; image upload endpoint will return error if invoked")
	}

	userService := service.NewUserService(userRepo, gotifyService)
	productService := service.NewProductService(productRepo, gotifyService, storageService)
	orderService := service.NewOrderService(orderRepo, productRepo, userRepo, gotifyService)
	analysisService := service.NewAnalysisService(analysisRepo)

	userHandler := handler.NewUserHandler(userService)
	productHandler := handler.NewProductHandler(productService)
	orderHandler := handler.NewOrderHandler(orderService)
	analysisHandler := handler.NewAnalysisHandler(analysisService)

	router := handler.SetupRouter(handler.RouterDependencies{
		UserHandler:     userHandler,
		ProductHandler:  productHandler,
		OrderHandler:    orderHandler,
		AnalysisHandler: analysisHandler,
	})

	addr := ":" + cfg.Port
	logger.Printf("Server listening on port %s...", cfg.Port)
	if err := http.ListenAndServe(addr, router); err != nil {
		logger.Fatalf("Server failed to start: %v", err)
	}
}
