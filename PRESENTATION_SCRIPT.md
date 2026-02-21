# 7-Minute Presentation Script

## Event Registration & Ticketing System
### Concurrency-Safe Registration with SELECT FOR UPDATE

---

## Minute 0-1: Introduction (1 minute)

> "Good [morning/afternoon/evening], everyone. I'm here to present my Event Registration & Ticketing System - a production-grade REST API built in Go that handles concurrent event registrations safely.
>
> The key challenge I solved: How do you prevent overbooking when 50 users try to register for an event with only 10 seats available?
>
> Let me show you how I solved this with database-level concurrency control."

---

## Minute 1-2: Architecture Overview (1 minute)

> "Here's the project structure following Clean Architecture:
>
> - **Handler Layer** - Receives HTTP requests, returns JSON responses
> - **Service Layer** - Business logic, including my concurrency control
> - **Repository Layer** - Database operations with GORM
> - **Model Layer** - User, Event, Registration data structures
>
> The request flow is simple: HTTP comes in at the top, flows down through each layer, and hits PostgreSQL at the bottom."

---

## Minute 2-3: The Problem (1 minute)

> "Let me explain the race condition problem.
>
> Without proper locking, here's what happens when two users try to register simultaneously:
>
> - Both read: seats = 1
> - Both check: seats > 0 ✓
> - Both insert registration records
> - Both decrement: seats = 0
>
> **Result: 2 registrations for 1 seat!** This is called a race condition."

---

## Minute 3-4: The Solution (1 minute)

> "My solution uses PostgreSQL's SELECT FOR UPDATE - a row-level lock.
>
> Here's the flow:
> 1. BEGIN TRANSACTION
> 2. SELECT * FROM events WHERE id = 1 **FOR UPDATE** ← Locks the row!
> 3. Check available_seats > 0
> 4. INSERT registration
> 5. UPDATE available_seats = available_seats - 1 WHERE seats > 0
> 6. COMMIT
>
> Other transactions wait at step 2 until the lock is released. This guarantees only one transaction can modify the seat count at a time."

---

## Minute 4-5: Demo (1 minute)

> "Let me show you the working system.
>
> [Switch to terminal/demo]
>
> - Server running on port 8080
> - Created an organizer user
> - Created an event with capacity 10
> - Now let me register users...
>
> [Show registration working, seats decrementing]
>
> Notice the server logs show the FOR UPDATE query and atomic seat decrement!"

---

## Minute 5-6: Concurrency Test (1 minute)

> "Now for the real test - I created a concurrency simulation with 50 goroutines.
>
> The test:
> - Creates an event with 10 seats
> - Launches 50 concurrent registration attempts
> - Only 10 should succeed
>
> [Run test or show results]
>
> **Exactly 10 succeeded, 40 failed gracefully with 'Event Full'** - no overbooking!"

---

## Minute 6-7: Summary & Key Takeaways (1 minute)

> "To summarize what I've built:
>
> ✅ A complete REST API with Go and Gin
> ✅ PostgreSQL database with GORM ORM
> ✅ Clean Architecture pattern
> ✅ **Concurrency-safe registration using SELECT FOR UPDATE**
> ✅ Tested with 50 simultaneous requests
> ✅ Proper error handling and HTTP status codes
>
> **The key learning**: Database-level locking (FOR UPDATE) is more reliable than application-level locks because it works even when multiple servers are running.
>
> Thank you! Questions?"
