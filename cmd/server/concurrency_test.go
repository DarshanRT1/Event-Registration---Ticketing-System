package main

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	"event-api/config"
	"event-api/models"
	"event-api/repository"
	"event-api/service"

	"gorm.io/gorm"
)

/*
ConcurrencyTest simulates 50 concurrent goroutines trying to register
for an event with capacity 10. Only 10 should succeed, and the rest
should fail gracefully with "event is full" error.

This test validates the concurrency-safe registration implementation.
*/
func ConcurrencyTest() {
	// Load configuration
	cfg := config.LoadConfig()

	// Connect to database
	db, err := cfg.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	eventRepo := repository.NewEventRepository(db)
	registrationRepo := repository.NewRegistrationRepository(db)

	// Initialize services
	userService := service.NewUserService(userRepo)
	eventService := service.NewEventService(eventRepo)
	registrationService := service.NewRegistrationService(db, eventRepo, registrationRepo, userRepo)

	// Setup test data
	setupConcurrencyTestData(db, userService, eventService)

	// Run the test
	runRegistrationTest(registrationService, eventRepo, db)
}

// setupConcurrencyTestData creates test users and an event with limited capacity
func setupConcurrencyTestData(db *gorm.DB, userService service.UserService, eventService service.EventService) {
	// Clean up previous test data
	db.Exec("DELETE FROM registrations")
	db.Exec("DELETE FROM events")
	db.Exec("DELETE FROM users WHERE email LIKE 'testuser%@example.com'")

	// Create organizer
	organizer := &models.User{
		Name:  "Test Organizer",
		Email: "organizer@example.com",
		Role:  models.RoleOrganizer,
	}

	// Check if organizer exists
	existingOrg, _ := userService.GetUserByEmail("organizer@example.com")
	if existingOrg == nil {
		if err := userService.CreateUser(organizer); err != nil {
			log.Printf("Warning: Could not create organizer: %v", err)
		}
	} else {
		organizer = existingOrg
	}

	// Create event with capacity 10
	event := &models.Event{
		Title:          "Concurrency Test Event",
		Capacity:       10,
		AvailableSeats: 10,
		OrganizerID:    organizer.ID,
	}

	// Check if event exists
	events, _ := eventService.GetEventsByOrganizerID(organizer.ID)
	var testEvent *models.Event
	for _, e := range events {
		if e.Title == "Concurrency Test Event" {
			testEvent = &e
			break
		}
	}

	if testEvent == nil {
		if err := eventService.CreateEvent(event); err != nil {
			log.Printf("Warning: Could not create event: %v", err)
		}
	} else {
		// Reset event seats
		event = testEvent
		db.Model(&models.Event{}).Where("id = ?", event.ID).Update("available_seats", 10)
	}

	// Create 50 test users
	for i := 1; i <= 50; i++ {
		user := &models.User{
			Name:  fmt.Sprintf("Test User %d", i),
			Email: fmt.Sprintf("testuser%d@example.com", i),
			Role:  models.RoleAttendee,
		}

		existingUser, _ := userService.GetUserByEmail(user.Email)
		if existingUser == nil {
			userService.CreateUser(user)
		}
	}

	log.Println("Test data setup complete: 1 event (capacity 10), 50 users")
}

// runRegistrationTest runs 50 concurrent registration attempts
func runRegistrationTest(registrationService service.RegistrationService, eventRepo repository.EventRepository, db *gorm.DB) {
	const (
		numGoroutines = 50
		eventID       = 1 // Will be updated dynamically
	)

	// Get the test event
	events, _ := eventRepo.FindAll()
	var testEvent *models.Event
	for _, e := range events {
		if e.Title == "Concurrency Test Event" {
			testEvent = &e
			break
		}
	}

	if testEvent == nil {
		log.Println("ERROR: Test event not found")
		return
	}

	// Get all test users
	users, _ := service.NewUserService(repository.NewUserRepository(db)).GetAllUsers()
	var testUsers []models.User
	for _, u := range users {
		if u.Email >= "testuser1@example.com" && u.Email <= "testuser50@example.com" {
			testUsers = append(testUsers, u)
		}
	}

	log.Printf("Starting concurrency test with %d goroutines for event ID %d (capacity: %d)",
		numGoroutines, testEvent.ID, testEvent.Capacity)

	// Counters for tracking results
	var successCount int32
	var failCount int32
	var eventFullCount int32
	var alreadyRegisteredCount int32

	// Use WaitGroup to wait for all goroutines
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Run 50 concurrent registration attempts
	for i := 0; i < numGoroutines && i < len(testUsers); i++ {
		go func(userID uint) {
			defer wg.Done()

			registration, err := registrationService.RegisterForEvent(userID, testEvent.ID)
			if err != nil {
				atomic.AddInt32(&failCount, 1)
				if err == models.ErrEventFull {
					atomic.AddInt32(&eventFullCount, 1)
				} else if err == models.ErrAlreadyRegistered {
					atomic.AddInt32(&alreadyRegisteredCount, 1)
				}
				log.Printf("Registration FAILED for user %d: %v", userID, err)
			} else {
				atomic.AddInt32(&successCount, 1)
				log.Printf("Registration SUCCESS for user %d: registration ID %d", userID, registration.ID)
			}
		}(testUsers[i].ID)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Print results
	log.Println("\n========== CONCURRENCY TEST RESULTS ==========")
	log.Printf("Total goroutines: %d", numGoroutines)
	log.Printf("Successful registrations: %d", atomic.LoadInt32(&successCount))
	log.Printf("Failed registrations: %d", atomic.LoadInt32(&failCount))
	log.Printf("  - Event full errors: %d", atomic.LoadInt32(&eventFullCount))
	log.Printf("  - Already registered errors: %d", atomic.LoadInt32(&alreadyRegisteredCount))

	// Verify final state
	updatedEvent, _ := eventRepo.FindByID(testEvent.ID)
	log.Printf("\nFinal event state:")
	log.Printf("  - Capacity: %d", updatedEvent.Capacity)
	log.Printf("  - Available seats: %d", updatedEvent.AvailableSeats)
	log.Printf("  - Registered: %d", updatedEvent.Capacity-updatedEvent.AvailableSeats)

	// Validate test results
	if atomic.LoadInt32(&successCount) == 10 && atomic.LoadInt32(&eventFullCount) >= 40 {
		log.Println("\n✅ TEST PASSED: Exactly 10 registrations succeeded, rest failed gracefully!")
	} else {
		log.Println("\n❌ TEST FAILED: Unexpected result!")
	}
	log.Println("================================================")
}

// RunTest is exported to be called from main
func RunTest() {
	ConcurrencyTest()
}
