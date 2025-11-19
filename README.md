# Vishwakarma Setu - Backend API 🏗️

The core backend microservice for **Vishwakarma Setu**, a trusted B2B marketplace for verifying, trading, and leasing industrial machinery.

This service manages **machine listings, rentals, search**, and acts as the **primary API** for the frontend.

---

## 🚀 Tech Stack

- **Language:** Go (Golang)
- **Framework:** Echo v4 – High-performance, extensible web framework
- **ORM:** GORM – Fantastic ORM for Go
- **Database:** PostgreSQL
- **Authentication:** JWT (via echo-jwt)

---

## 🛠️ Prerequisites

Ensure the following are installed:

- Go (1.25.4)
- PostgreSQL (local or Docker)
- Git

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

```env
# Server Config
PORT=1326
FRONTEND_URL=http://localhost:3000

# Database Config
DATABASE_DSN="host=localhost user=vishwakarma_user password=password dbname=vishwakarma_db port=5432 sslmode=disable"

# Auth Config
JWT_SECRET="your_jwt_secret_key"
```

### 4. Database Setup

Ensure PostgreSQL is running, then create the database:

```sql
CREATE DATABASE vishwakarma_db;
```

> The application will automatically migrate schema on first run.

---

## 🏃‍♂️ Running the Application

### Development Mode

```bash
go run .
```

The server runs at **[http://localhost:1326](http://localhost:1326)** (or the configured `PORT`).

---

## 📚 API Endpoints

---

## 🟢 Public Routes

These routes **do not require authentication**.

| Method | Endpoint            | Description                       |
| ------ | ------------------- | --------------------------------- |
| GET    | `/health`       | Check server status               |
| GET    | `/api/machines`     | Get all machine listings          |
| GET    | `/api/machines/:id` | Get details of a specific machine |

### Query Parameters for `/api/machines`

| Query | Description                             | Example                   |
| ----- | --------------------------------------- | ------------------------- |
| type  | Filter by listing type (`sale`, `rent`) | `/api/machines?type=rent` |

---

## 🔒 Protected Routes

These routes **require JWT authentication**.

**Header format:**

```
Authorization: Bearer <your_jwt_token>
```

| Method | Endpoint            | Description                  | Body |
| ------ | ------------------- | ---------------------------- | ---- |
| POST   | `/api/machines`     | Create a new machine listing | JSON |
| PUT    | `/api/machines/:id` | Update an existing listing   | JSON |
| DELETE | `/api/machines/:id` | Delete a listing             | None |

---

### 📌 Create Machine Example

**POST** `/api/machines`

```json
{
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
}
```

---

## 📂 Project Structure

```
vishwakarma-setu-backend/
├── config/
│   └── database.go      # DB connection & migration logic
├── controllers/
│   └── listing.go       # Handlers for Machine CRUD operations
├── middleware/
│   └── auth.go          # JWT validation middleware configuration
├── models/
│   └── machine.go       # Gorm model definition for 'machines'
├── routes/
│   └── routes.go        # API route definitions & grouping
├── .env                 # Environment variables (ignored in Git)
├── go.mod               # Go module definition
└── main.go              # Entry point & server configuration
```

---

## 🐞 Troubleshooting

### ❌ **"Invalid token claims"**

* Ensure the Auth service includes `user_id` (numeric) in JWT payload.
* Verify `JWT_SECRET` matches exactly in both services.

---

### ❌ Database connection failed

* Check if PostgreSQL is running.
* Verify `.env` credentials.
* Ensure the database exists.