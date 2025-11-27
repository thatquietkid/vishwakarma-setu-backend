# Vishwakarma Setu - Backend API 🏗️

The core backend microservice for Vishwakarma Setu, a trusted B2B marketplace for verifying, trading, and leasing industrial machinery.

This service manages machine listings, rentals, inspections, maintenance history, and payments, acting as the primary API for the frontend.

---

## 📋 Table of Contents

- [Tech Stack](#-tech-stack) 
- [Prerequisites](#-prerequisites)
- [Setup & Installation](#-setup--installation)
- [Running the Application](#-running-the-application)
- [Testing & Coverage](#-testing--coverage)
- [API Documentation (Swagger)](#-api-documentation-swagger--)
- [API Endpoints & Examples](#-api-endpoints--examples-)
- [Project Structure](#-project-structure-)
- [Troubleshooting](#-troubleshooting-)

---

## 🚀 Tech Stack

- **Language:** Go (Golang)  
- **Framework:** Echo v4 – High-performance, extensible web framework  
- **ORM:** GORM – Fantastic ORM for Go  
- **Database:** PostgreSQL  
- **Authentication:** JWT (via echo-jwt)  
- **Documentation:** Swagger (Swaggo)

---

## 🛠️ Prerequisites

Ensure the following are installed:

- Go (1.25.4)  
- PostgreSQL (local or Docker)  
- Git  
- (Optional) Docker & Docker Compose  
- (Optional) Make (for running tests)

---

## ⚙️ Setup & Installation

### 1. Clone the repository

```bash
git clone https://github.com/thatquietkid/vishwakarma-setu-backend.git
cd vishwakarma-setu-backend
```

### 2. Install Dependencies

```bash
go mod tidy
```

### 3. Environment Configuration

Create a `.env` file in the project root:

```
# Server Config
PORT=1326
FRONTEND_URL=http://localhost:3000

# Database Config
# Update credentials as per your local setup
DATABASE_DSN="host=localhost user=vishwakarma_user password=password dbname=vishwakarma_db port=5432 sslmode=disable TimeZone=Asia/Kolkata"

# Auth Config (Must match Auth Service)
JWT_SECRET="your_jwt_secret_key"
```

---

## 🏃‍♂️ Running the Application

### Local Development

```bash
go run .
```

The server runs at **[http://localhost:1326](http://localhost:1326)** (or the configured PORT).
The API is mounted under the `/api` base path.

### Docker

To run the entire stack (Backend + Database) using Docker Compose:

```bash
docker-compose up --build
```

---

## 🧪 Testing & Coverage

We use gotestsum for formatted test output. A Makefile is provided for convenience.

### Run Tests

```bash
# Run all tests with status output
make test

# Run tests with detailed verbose output
make test-verbose
```

### Check Coverage

```bash
# Run tests and display coverage summary in terminal
make test-cover

# Generate and open detailed HTML coverage report in browser
make test-html
```

### Current Coverage Metrics:

* Controllers: ~89% Coverage
* Auth Logic: Verified via Integration Tests
* Rental Logic: Verified via Integration Tests

---

## 📘 API Documentation (Swagger / OpenAPI)

The project includes Swagger annotations.

### Install Swag:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

### Generate Docs:

```bash
swag init
```

### View Docs

Start the server and visit:
**[http://localhost:1326/swagger/index.html](http://localhost:1326/swagger/index.html)**

---

## 📚 API Endpoints & Examples

---

### 🟢 Public Routes

| Method | Endpoint                     | Description                             |
| ------ | ---------------------------- | --------------------------------------- |
| GET    | /health                      | Server health check                     |
| GET    | /api/machines                | Get all listings (Search, Filter, Sort) |
| GET    | /api/machines/:id            | Get machine details                     |
| GET    | /api/machines/:id/inspection | Get inspection report                   |

---

### 🔒 Protected Routes (Requires Bearer Token)

Header:
`Authorization: Bearer <your_jwt_token>`

---

### 1. Create Machine Listing

**Request:**

```bash
curl -X POST http://localhost:1326/api/machines \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "2019 Haas VF-2 CNC Mill",
    "description": "Excellent condition vertical machining center.",
    "manufacturer": "Haas Automation",
    "model_number": "VF-2",
    "year_of_manufacture": 2019,
    "listing_type": "both",
    "status": "listed",
    "price_for_sale": 2500000,
    "rental_price_per_month": 120000,
    "security_deposit": 200000,
    "specs": {
        "spindle_speed": "8100 RPM",
        "axis_travel": "30x16x20 inches"
    }
}'
```

**Response (201 Created):**

```json
{
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "seller_id": 1,
    "title": "2019 Haas VF-2 CNC Mill",
    "status": "listed",
    "created_at": "2025-11-27T10:00:00Z"
}
```

---

### 2. Book a Rental

**Request:**

```bash
curl -X POST http://localhost:1326/api/rentals \
  -H "Authorization: Bearer <TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "machine_id": "550e8400-e29b-41d4-a716-446655440000",
    "start_date": "2025-01-01",
    "end_date": "2025-01-05"
}'
```

**Response (201 Created):**

```json
{
    "id": "rental-uuid-1234",
    "machine_id": "550e8400-e29b-41d4-a716-446655440000",
    "total_amount": 20000,
    "security_deposit": 200000,
    "status": "pending"
}
```

---

## 📂 Project Structure

```
vishwakarma-setu-backend/
├── config/
│   └config.go           # DB configuration
├── controllers/
│   ├── index.go         # General helpers
│   ├── listing.go       # Machine CRUD
│   ├── rentals.go       # Rental logic
│   ├── inspection.go    # Inspection reports
│   ├── maintenance.go   # Maintenance history
│   ├── payment.go       # Payment gateway integration
│   └── upload.go        # File upload handler
├── middleware/
│   └── auth.go          # JWT Middleware
├── models/
│   ├── machine.go       # Machine schema
│   ├── rental.go        # Rental schema
│   ├── inspection.go    # Inspection schema
│   └── maintenance.go   # MaintenRoute definitions
├── scripts/
│   └── init_db.sh       # DB Init script
├── docs/                # Generated Swagger docs
├── .env                 # Env variables
├── docker-compose.yml   # Docker config
├── Dockerfile           # Docker build definition
├── Makefile             # Test commands
└── main.go              # Entry point
```
