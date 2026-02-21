# PowerPoint Presentation Outline
## Event Registration & Ticketing System

---

## Slide 1: Title Slide

**Title:** Event Registration & Ticketing System
**Subtitle:** Concurrency-Safe Registration with Go & PostgreSQL
**Presenter:** [Your Name]
**Date:** [Date]

---

## Slide 2: Problem Statement

**Title:** The Overbooking Problem

- What happens when 50 users try to register for 10 seats?
- Race condition: multiple reads → multiple writes → overbooking
- Traditional fixes: locks, semaphores (application-level)
- Our solution: Database-level locking

---

## Slide 3: Solution Overview

**Title:** SELECT FOR UPDATE

- PostgreSQL feature: Row-level locking
- Other transactions wait until lock released
- Guarantees atomic seat management

---

## Slide 4: Architecture

**Title:** Clean Architecture

```
┌────────────────────────────────────┐
│         HTTP (Gin Framework)       │
└──────────────┬─────────────────────┘
               │
┌──────────────▼─────────────────────┐
│            Handler Layer           │
└──────────────┬─────────────────────┘
               │
┌──────────────▼─────────────────────┐
│            Service Layer           │
│    (Business Logic + Concurrency)  │
└──────────────┬─────────────────────┘
               │
┌──────────────▼─────────────────────┐
│          Repository Layer          │
│         (GORM + PostgreSQL)        │
└────────────────────────────────────┘
```

---

## Slide 5: Project Structure

**Title:** Files & Folders

- `cmd/server/` - Entry point + tests
- `config/` - Database configuration
- `models/` - User, Event, Registration
- `repository/` - Data access layer
- `service/` - Business logic
- `handler/` - HTTP endpoints

---

## Slide 6: Database Schema

**Title:** Three Tables

1. **Users** - id, name, email, role
2. **Events** - id, title, capacity, available_seats, organizer_id
3. **Registrations** - id, user_id, event_id (unique constraint)

---

## Slide 7: Concurrency Flow

**Title:** Registration Flow with Locking

1. BEGIN TRANSACTION
2. SELECT ... FOR UPDATE (lock row)
3. CHECK available_seats > 0
4. INSERT registration
5. UPDATE available_seats - 1
6. COMMIT

---

## Slide 8: Code Example

**Title:** The Critical Code

```go
// In registration_service.go
tx := db.Begin()

// Lock the row
event := eventRepo.FindByIDForUpdate(tx, eventID)

// Check seats
if event.AvailableSeats <= 0 {
    tx.Rollback()
    return ErrEventFull
}

// Insert registration
registrationRepo.Create(tx, registration)

// Decrement seats
db.Model(&Event{}).Where("id = ? AND available_seats > 0", eventID).
    Update("available_seats", gorm.Expr("available_seats - 1"))

tx.Commit()
```

---

## Slide 9: Test Results

**Title:** 50 Goroutines → 10 Seats

- **Total attempts:** 50
- **Successful:** 10 ✅
- **Failed (event full):** 40 ✅
- **No overbooking!**

---

## Slide 10: Demo

**Title:** Live Demo

- Show API running
- Create user, create event
- Register and watch seats decrement

---

## Slide 11: Key Features

**Title:** What We Built

- ✅ REST API with Go + Gin
- ✅ PostgreSQL + GORM
- ✅ Clean Architecture
- ✅ Concurrency-safe registration
- ✅ Environment configuration
- ✅ Comprehensive error handling

---

## Slide 12: Tech Stack

**Title:** Technologies Used

| Layer | Technology |
|-------|------------|
| Language | Go 1.20+ |
| Web Framework | Gin |
| Database | PostgreSQL |
| ORM | GORM |

---

## Slide 13: Future Enhancements

**Title:** What's Next

- JWT Authentication
- Email notifications
- Payment integration
- Event categories
- Waitlist for full events

---

## Slide 14: Questions

**Title:** Q&A

- Thank you!
- Questions?

---

## Slide 15: References

**Title:** Resources

- Repository: github.com/DarshanRT1/Event-Registration---Ticketing-System
- Go: go.dev
- GORM: gorm.io
- PostgreSQL: postgresql.org

---

## Presentation Tips

- **Total Slides:** 15
- **Suggested Time:** 7-10 minutes
- **Key Point:** Emphasize the SELECT FOR UPDATE solution
- **Demo:** Live demo creates impact
- **Backup:** Have screenshots ready if demo fails
