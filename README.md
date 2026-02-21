# Event Registration & Ticketing System

A production-grade REST API built in Go for an Event Registration & Ticketing System featuring **concurrency-safe event registration** using database-level row locking (`SELECT FOR UPDATE`).

---

## 📋 Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
- [Technology Stack](#technology-stack)
- [Database Schema](#database-schema)
- [Quick Start](#quick-start)
- [API Reference](#api-reference)
- [Concurrency Strategy](#concurrency-strategy)
- [Running the Concurrency Test](#running-the-concurrency-test)
- [Sample HTTP Requests](#sample-http-requests)
- [Contributing](#contributing)
- [License](#license)

---

## Overview

This project implements a complete Event Registration & Ticketing System where:

- **Attendees** can browse events and register for them
- **Organizers** can create events with defined capacity limits
- The system **prevents overbooking** through database-level concurrency control

### The Core Challenge

When 50 users simultaneously try to register for an event with only 10 seats, race conditions can cause overbooking. This system solves that with **database transactions + row-level locking**.

---

## Features

- ✅ RESTful API with Gin framework
- ✅ PostgreSQL database with GORM ORM
- ✅ Clean Architecture (Handler → Service → Repository → Model)
- ✅ Environment variable configuration
- ✅ Concurrency-safe registration using `SELECT FOR UPDATE`
- ✅ Atomic seat management with transactional integrity
- ✅ Comprehensive error handling with appropriate HTTP status codes
- ✅ Concurrency simulation test (50 goroutines → 10 seats)

---

## Architecture

### Request Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           HTTP REQUEST                                      │
│                         (curl / Postman)                                   │
└─────────────────────────────────┬───────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                            HANDLER LAYER                                   │
│                    (handler/user_handler.go)                                │
│                 handler/event_handler.go                                    │
│            handler/registration_handler.go                                  │
│   • Parses HTTP requests                                                   │
│   • Validates input                                                        │
│   • Returns JSON responses                                                 │
└─────────────────────────────────┬───────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           SERVICE LAYER                                     │
│                   (service/user_service.go)                                │
│                    service/event_service.go                                 │
│              service/registration_service.go                                │
│   • Business logic                                                         │
│   • Transaction management (SELECT FOR UPDATE)                              │
│   • Concurrency control                                                    │
└─────────────────────────────────┬───────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         REPOSITORY LAYER                                   │
│                 (repository/user_repository.go)                             │
│              repository/event_repository.go                                 │
│          repository/registration_repository.go                              │
│   • Database operations                                                    │
│   • GORM queries                                                           │
│   • Raw SQL when needed                                                    │
└─────────────────────────────────┬───────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                          DATABASE LAYER                                    │
│                         (PostgreSQL)                                       │
│   • Users table                                                            │
│   • Events table                                                           │
│   • Registrations table                                                    │
│   • Row-level locking (FOR UPDATE)                                         │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Project Structure

```
event-registration-ticketing-system/
├── cmd/
│   └── server/
│       ├── main.go                   # Application entry point
│       └── concurrency_test.go       # Concurrency simulation (50 goroutines)
├── config/
│   └── config.go                    # Configuration & DB connection
├── models/
│   └── models.go                    # User, Event, Registration models
├── repository/
│   ├── user_repository.go           # User data access
│   ├── event_repository.go          # Event data access (includes FOR UPDATE)
│   └── registration_repository.go    # Registration data access
├── service/
│   ├── user_service.go              # User business logic
│   ├── event_service.go             # Event business logic
│   └── registration_service.go      # Core concurrency-safe registration
├── handler/
│   ├── user_handler.go              # User HTTP endpoints
│   ├── event_handler.go             # Event HTTP endpoints
│   └── registration_handler.go      # Registration HTTP endpoints
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

---

## Technology Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.20+ |
| Web Framework | Gin |
| Database | PostgreSQL 12+ |
| ORM | GORM |
| Architecture | Clean Architecture |

---

## Database Schema

### Users Table
```sql
CREATE TABLE users (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    email       VARCHAR(255) UNIQUE NOT NULL,
    role        VARCHAR(50) DEFAULT 'attendee',
    created_at  TIMESTAMP,
    updated_at  TIMESTAMP,
    deleted_at  TIMESTAMP
);
```

### Events Table
```sql
CREATE TABLE events (
    id              SERIAL PRIMARY KEY,
    title           VARCHAR(255) NOT NULL,
    capacity        INTEGER NOT NULL,
    available_seats INTEGER NOT NULL,
    organizer_id    INTEGER REFERENCES users(id),
    created_at      TIMESTAMP,
    updated_at      TIMESTAMP,
    deleted_at      TIMESTAMP
);
```

### Registrations Table
```sql
CREATE TABLE registrations (
    id          SERIAL PRIMARY KEY,
    user_id     INTEGER REFERENCES users(id),
    event_id    INTEGER REFERENCES events(id),
    created_at  TIMESTAMP,
    updated_at  TIMESTAMP,
    deleted_at  TIMESTAMP,
    UNIQUE(user_id, event_id)
);
```

---

## Quick Start

### Prerequisites

- Go 1.20 or later
- PostgreSQL 12 or later

### 1. Clone & Setup

```bash
# Clone the repository
git clone https://github.com/DarshanRT1/Event-Registration---Ticketing-System.git
cd Event-Registration---Ticketing-System

# Initialize Go module (if not already)
go mod tidy

# Download dependencies
go mod download
```

### 2. Configure Database

Create a `.env` file (optional):

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=eventdb
SERVER_PORT=8080
```

Or set environment variables:

```bash
# Windows
set DB_PASSWORD=your_password
go run cmd/server/main.go

# Linux/Mac
export DB_PASSWORD=your_password
go run cmd/server/main.go
```

### 3. Run the Server

```bash
go run cmd/server/main.go
```

Expected output:
```
Database connection established and migrations completed
Server starting on :8080
```

---

## API Reference

### Base URL

```
http://localhost:8080
```

### Endpoints

#### Users

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/users` | Create a new user |
| GET | `/api/v1/users` | Get all users |
| GET | `/api/v1/users/:id` | Get user by ID |
| PUT | `/api/v1/users/:id` | Update user |
| DELETE | `/api/v1/users/:id` | Delete user |

#### Events

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/events` | Create a new event |
| GET | `/api/v1/events` | Get all events |
| GET | `/api/v1/events/:id` | Get event by ID |
| PUT | `/api/v1/events/:id` | Update event |
| DELETE | `/api/v1/events/:id` | Delete event |
| GET | `/api/v1/events/organizer/:organizerID` | Get events by organizer |

#### Registrations

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/registrations` | Register for an event |
| GET | `/api/v1/registrations/:id` | Get registration by ID |
| GET | `/api/v1/registrations/user/:userID` | Get user's registrations |
| GET | `/api/v1/registrations/event/:eventID` | Get event's registrations |
| DELETE | `/api/v1/registrations` | Cancel registration |

---

## Concurrency Strategy

### The Problem

Without proper concurrency control, this race condition occurs:

```
Time    Thread A                   Thread B
─────────────────────────────────────────────────
T1      READ seats=1               
T2                               READ seats=1
T3      CHECK seats>0 ✓            
T3                               CHECK seats>0 ✓
T4      INSERT registration       
T5                               INSERT registration
T6      seats = 1-1 = 0           
T7                               seats = 1-1 = 0
─────────────────────────────────────────────────
RESULT: 2 registrations, 0 seats! ❌ (OVERBOOKING)
```

### Our Solution: SELECT FOR UPDATE

We use **database transactions with row-level locking**:

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     CONCURRENCY CONTROL FLOW                            │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  1. BEGIN TRANSACTION                                                   │
│         │                                                               │
│         ▼                                                               │
│  2. SELECT * FROM events WHERE id = ? FOR UPDATE                       │
│     🔒 Row is now LOCKED - other transactions WAIT here                 │
│         │                                                               │
│         ▼                                                               │
│  3. CHECK available_seats > 0                                          │
│     If false → ROLLBACK → "Event Full"                                 │
│         │                                                               │
│         ▼                                                               │
│  4. INSERT INTO registrations (user_id, event_id)                       │
│         │                                                               │
│         ▼                                                               │
│  5. UPDATE events SET available_seats = available_seats - 1             │
│     WHERE id = ? AND available_seats > 0                                │
│     🔒 Atomic decrement with safety check                                │
│         │                                                               │
│         ▼                                                               │
│  6. COMMIT → Lock released                                              │
│                                                                          │
│  On ANY Error: ROLLBACK (cancels all changes)                           │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Key Mechanisms

1. **`SELECT ... FOR UPDATE`**: Acquires an exclusive row lock
2. **Atomic UPDATE**: `available_seats - 1` with `WHERE available_seats > 0`
3. **Rows Affected Check**: Verifies the UPDATE actually modified a row
4. **Unique Constraint**: `(user_id, event_id)` prevents duplicate registrations

---

## Running the Concurrency Test

### How It Works

The test simulates 50 concurrent users trying to register for an event with only 10 seats.

```bash
# Run the test (call from main or separate terminal)
go run cmd/server/main.go
# Then use curl to trigger concurrent registrations
```

### Expected Results

```
========== CONCURRENCY TEST RESULTS ==========
Total goroutines: 50
Successful registrations: 10
Failed registrations: 40
  - Event full errors: 40
  - Already registered errors: 0

Final event state:
  - Capacity: 10
  - Available seats: 0
  - Registered: 10

✅ TEST PASSED: Exactly 10 registrations succeeded!
================================================
```

---

## Sample HTTP Requests

### Using cURL

#### 1. Health Check
```bash
curl http://localhost:8080/health
```
Response: `{"status":"ok"}`

#### 2. Create Organizer User
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Organizer",
    "email": "john@example.com",
    "role": "organizer"
  }'
```

#### 3. Create Attendee User
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Jane Attendee",
    "email": "jane@example.com",
    "role": "attendee"
  }'
```

#### 4. Create Event
```bash
curl -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Go Conference 2024",
    "capacity": 100,
    "organizer_id": 1
  }'
```

#### 5. Get All Events
```bash
curl http://localhost:8080/api/v1/events
```

#### 6. Register for Event
```bash
curl -X POST http://localhost:8080/api/v1/registrations \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 2,
    "event_id": 1
  }'
```

#### 7. Get User Registrations
```bash
curl http://localhost:8080/api/v1/registrations/user/2
```

#### 8. Cancel Registration
```bash
curl -X DELETE http://localhost:8080/api/v1/registrations \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 2,
    "event_id": 1
  }'
```

### Using Postman

Import the following collection:

```json
{
  "info": {
    "name": "Event Registration API",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Create User",
      "request": {
        "method": "POST",
        "url": "http://localhost:8080/api/v1/users",
        "body": {
          "mode": "raw",
          "raw": "{\"name\":\"John\",\"email\":\"john@test.com\",\"role\":\"organizer\"}"
        }
      }
    },
    {
      "name": "Create Event",
      "request": {
        "method": "POST",
        "url": "http://localhost:8080/api/v1/events",
        "body": {
          "mode": "raw", 
          "raw": "{\"title\":\"Tech Talk\",\"capacity\":10,\"organizer_id\":1}"
        }
      }
    },
    {
      "name": "Register",
      "request": {
        "method": "POST",
        "url": "http://localhost:8080/api/v1/registrations",
        "body": {
          "mode": "raw",
          "raw": "{\"user_id\":2,\"event_id\":1}"
        }
      }
    }
  ]
}
```

---

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## License

MIT License - see [LICENSE](LICENSE) for details.

---

**Built with ❤️ using Go, Gin, PostgreSQL, and GORM**
