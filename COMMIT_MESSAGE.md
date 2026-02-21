# Commit & PR Messages

## Commit Message

```
feat: implement concurrency-safe event registration with SELECT FOR UPDATE

- Add PostgreSQL database with GORM ORM
- Implement Clean Architecture (handlers→services→repository→models)
- Add concurrency control using database row-level locking
- Include 50-goroutine concurrency test demonstrating seat reservation
- Add comprehensive README with architecture diagrams and setup guide

Key implementation:
- SELECT FOR UPDATE in event_repository.go locks rows during registration
- Atomic seat decrement with WHERE available_seats > 0 clause
- Transaction rollback on any error ensures data consistency
- Tested: 50 concurrent registrations → exactly 10 succeed
```

---

## PR Description Template

```markdown
## Summary

This PR implements a production-grade Event Registration & Ticketing System in Go with concurrency-safe seat reservation.

## Problem Solved

When multiple users try to register for the same event simultaneously, race conditions can cause overbooking. This solution uses PostgreSQL's `SELECT FOR UPDATE` to lock the event row during registration, ensuring only one transaction can modify available seats at a time.

## Key Features

- ✅ REST API with Gin framework
- ✅ PostgreSQL database with GORM ORM
- ✅ Clean Architecture pattern
- ✅ Concurrency-safe registration using database-level locking
- ✅ Comprehensive test: 50 goroutines → exactly 10 succeed
- ✅ Environment variable configuration

## Architecture

```
HTTP Request → Handler → Service → Repository → PostgreSQL
                      ↓
              SELECT FOR UPDATE (row lock)
```

## Changes

| File | Description |
|------|-------------|
| `cmd/server/main.go` | Application entry point |
| `cmd/server/concurrency_test.go` | 50-goroutine test |
| `service/registration_service.go` | Core concurrency logic |
| `repository/event_repository.go` | FOR UPDATE implementation |
| `config/config.go` | DB configuration |
| `README.md` | Complete documentation |

## Testing

```
========== CONCURRENCY TEST RESULTS ==========
Total goroutines: 50
Successful registrations: 10
Failed registrations: 40 (event full)
✅ TEST PASSED: Exactly 10 registrations succeeded!
```

## How to Run

```bash
# Set database credentials
export DB_PASSWORD=your_password

# Run server
go run cmd/server/main.go

# Test endpoints
curl http://localhost:8080/api/v1/events
```

## Related Issues

- Fixes #1: Prevent event overbooking
- Implements concurrency control best practices
```

---

## Quick Copy for Git

```bash
git add -A
git commit -m "feat: implement concurrency-safe event registration with SELECT FOR UPDATE"
git push origin main
```

Then create PR using the template above.
