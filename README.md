# PostgreSQL Repository Pattern API with Analytics & Gotify Integration

A clean, modular REST API built in Go 1.22 using the **Repository Pattern** and **Clean Architecture**. It manages 3 core PostgreSQL database entities (**User**, **Product**, **Order**), provides **Analytics/Analysis endpoints**, and triggers **Gotify** push notifications.

---

## Architecture Overview

- **Domain Layer** (`internal/domain`): Core models and interfaces for entities and repositories.
- **Repository Layer** (`internal/repository/postgres`): PostgreSQL implementations with raw SQL queries and transactions.
- **Service Layer** (`internal/service`): Business logic, stock deduction validations, and Gotify event dispatchers.
- **Notification Layer** (`internal/notification`): Gotify HTTP notification client.
- **Handler Layer** (`internal/handler`): HTTP REST controllers utilizing Go 1.22 enhanced routing.

---

## Features

- **PostgreSQL Database**: Auto-initializes schema (`users`, `products`, `orders`, `order_items`).
- **User Management**: Full CRUD (`/api/v1/users`).
- **Product Management**: Full CRUD (`/api/v1/products`) with stock quantity tracking.
- **Order Management**: Order placement with stock deduction, order item mapping, total computation, and status updates (`/api/v1/orders`).
- **Analytics & Analysis**:
  - `GET /api/v1/analysis/summary` - Total users, products, orders, revenue, average order value, low-stock count.
  - `GET /api/v1/analysis/top-products` - Top selling products by quantity and revenue.
  - `GET /api/v1/analysis/user-stats` - Top spending customers.
  - `GET /api/v1/analysis/sales-trend` - Orders & revenue breakdown by status.
- **Gotify Push Notifications**:
  - Triggers alerts on User Registration, New Order placement, Order Status updates, and Low Stock (< 5 items) warnings.

---

## Running with Docker Compose

Start the Go API, PostgreSQL, and Gotify services:

```bash
docker-compose up --build
```

- **API Endpoint**: `http://localhost:8080`
- **PostgreSQL Database**: `localhost:5432`
- **Gotify Server Dashboard**: `http://localhost:8088` (Default login: `admin` / `admin`)

---

## API Endpoints

### 1. Users
- `POST /api/v1/users` - Create user
- `GET /api/v1/users` - List all users
- `GET /api/v1/users/{id}` - Get user details
- `PUT /api/v1/users/{id}` - Update user
- `DELETE /api/v1/users/{id}` - Delete user

### 2. Products
- `POST /api/v1/products` - Create product
- `GET /api/v1/products` - List products
- `GET /api/v1/products/{id}` - Get product details
- `PUT /api/v1/products/{id}` - Update product
- `DELETE /api/v1/products/{id}` - Delete product

### 3. Orders
- `POST /api/v1/orders` - Place new order
- `GET /api/v1/orders` - List all orders (filter with `?user_id=1`)
- `GET /api/v1/orders/{id}` - Get order details with item list
- `PATCH /api/v1/orders/{id}/status` - Update status (`pending`, `completed`, `cancelled`)
- `DELETE /api/v1/orders/{id}` - Cancel/Delete order

### 4. Analysis
- `GET /api/v1/analysis/summary` - Overall metrics summary
- `GET /api/v1/analysis/top-products?limit=5` - Most sold products
- `GET /api/v1/analysis/user-stats?limit=5` - Top spending customers
- `GET /api/v1/analysis/sales-trend` - Sales distribution by status
