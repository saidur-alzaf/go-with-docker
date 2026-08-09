package main

import (
	"log"
	"net/http"
	"os"

	"go-sqlite-api/internal/config"
	"go-sqlite-api/internal/database"
	"go-sqlite-api/internal/handler"
	"go-sqlite-api/internal/notification"
	"go-sqlite-api/internal/repository/postgres"
	"go-sqlite-api/internal/service"
)

func main() {
	logger := log.New(os.Stdout, "[APP] ", log.LstdFlags|log.Lshortfile)

	// 1. Load Configuration
	cfg := config.LoadConfig()

	// 2. Connect to separate PostgreSQL databases
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

	// 3. Run per-entity migrations
	if err := database.MigrateUserDB(userDB); err != nil {
		logger.Fatalf("User DB migration error: %v", err)
	}
	if err := database.MigrateProductDB(productDB); err != nil {
		logger.Fatalf("Product DB migration error: %v", err)
	}
	if err := database.MigrateOrderDB(orderDB); err != nil {
		logger.Fatalf("Order DB migration error: %v", err)
	}

	// 4. Initialize Repositories
	userRepo := postgres.NewUserRepository(userDB)
	productRepo := postgres.NewProductRepository(productDB)
	orderRepo := postgres.NewOrderRepository(orderDB)
	analysisRepo := postgres.NewAnalysisRepository(userDB, productDB, orderDB)

	// 5. Initialize Notification Service (Gotify)
	gotifyService := notification.NewGotifyService(cfg.GotifyURL, cfg.GotifyToken)

	// 6. Initialize Services
	userService := service.NewUserService(userRepo, gotifyService)
	productService := service.NewProductService(productRepo, gotifyService)
	orderService := service.NewOrderService(orderRepo, productRepo, userRepo, gotifyService)
	analysisService := service.NewAnalysisService(analysisRepo)

	// 7. Initialize Handlers
	userHandler := handler.NewUserHandler(userService)
	productHandler := handler.NewProductHandler(productService)
	orderHandler := handler.NewOrderHandler(orderService)
	analysisHandler := handler.NewAnalysisHandler(analysisService)

	// 8. Setup Router
	router := handler.SetupRouter(handler.RouterDependencies{
		UserHandler:     userHandler,
		ProductHandler:  productHandler,
		OrderHandler:    orderHandler,
		AnalysisHandler: analysisHandler,
	})

	// 9. Start HTTP Server
	addr := ":" + cfg.Port
	logger.Printf("Server listening on port %s...", cfg.Port)
	if err := http.ListenAndServe(addr, router); err != nil {
		logger.Fatalf("Server failed to start: %v", err)
	}
}
