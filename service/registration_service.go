package service

import (
	"event-api/models"
	"event-api/repository"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// RegistrationService handles registration business logic
type RegistrationService interface {
	RegisterForEvent(userID, eventID uint) (*models.Registration, error)
	GetRegistrationByID(id uint) (*models.Registration, error)
	GetUserRegistrations(userID uint) ([]models.Registration, error)
	GetEventRegistrations(eventID uint) ([]models.Registration, error)
	CancelRegistration(userID, eventID uint) error
}

type registrationService struct {
	db               *gorm.DB
	eventRepo        repository.EventRepository
	registrationRepo repository.RegistrationRepository
	userRepo         repository.UserRepository
}

// NewRegistrationService creates a new RegistrationService
// This is the core service that handles concurrency-safe event registration
func NewRegistrationService(
	db *gorm.DB,
	eventRepo repository.EventRepository,
	registrationRepo repository.RegistrationRepository,
	userRepo repository.UserRepository,
) RegistrationService {
	return &registrationService{
		db:               db,
		eventRepo:        eventRepo,
		registrationRepo: registrationRepo,
		userRepo:         userRepo,
	}
}

/*
RegisterForEvent implements the critical concurrency-safe registration logic.

CONCURRENCY STRATEGY:
=====================

The registration must be atomic to prevent overbooking. Here's the step-by-step process:

1. BEGIN TRANSACTION - Start a database transaction to ensure atomicity
2. SELECT FOR UPDATE - Lock the event row to prevent other transactions from modifying it
3. CHECK AVAILABLE SEATS - Verify that available_seats > 0
4. INSERT REGISTRATION - Add the registration record
5. DECREMENT SEATS - Atomically decrease available_seats by 1
6. COMMIT - Save all changes or ROLLBACK on any error

Why this works:
- The SELECT FOR UPDATE clause locks the row until the transaction completes
- Other concurrent transactions will wait at step 2 until the lock is released
- This ensures only one transaction can modify the seats count at a time
- The WHERE clause in step 5 (available_seats > 0) provides an additional safety net
- If any step fails, the entire transaction is rolled back

This approach prevents race conditions like:
- Multiple goroutines reading available_seats = 1 simultaneously
- Multiple goroutines inserting registrations
- Overbooking due to concurrent seat decrements
*/
func (s *registrationService) RegisterForEvent(userID, eventID uint) (*models.Registration, error) {
	// Validate user exists
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrUserNotFound
		}
		return nil, err
	}

	// Start a new database transaction
	// All operations within this transaction will be atomic
	tx := s.db.Begin()

	// Check if user is already registered (within transaction)
	var existingReg models.Registration
	err = tx.Where("user_id = ? AND event_id = ?", userID, eventID).First(&existingReg).Error
	if err == nil {
		// User already registered - rollback and return error
		tx.Rollback()
		return nil, models.ErrAlreadyRegistered
	}
	if err != gorm.ErrRecordNotFound {
		tx.Rollback()
		return nil, err
	}

	// CRITICAL: Lock the event row using SELECT FOR UPDATE
	// This prevents other transactions from modifying this row until we commit/rollback
	event, err := s.eventRepo.FindByIDForUpdate(tx, eventID)
	if err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return nil, models.ErrEventNotFound
		}
		return nil, err
	}

	// CRITICAL: Check if seats are available
	// This check happens AFTER acquiring the lock, so it's safe
	if event.AvailableSeats <= 0 {
		tx.Rollback()
		return nil, models.ErrEventFull
	}

	// Create the registration record
	registration := &models.Registration{
		UserID:  userID,
		EventID: eventID,
	}

	// Use ON CONFLICT to handle race condition on unique constraint
	// Even though we checked above, this provides defense in depth
	err = tx.Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(registration).Error

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// CRITICAL: Atomically decrement available seats
	// Use UPDATE with WHERE clause for additional safety
	result := tx.Model(&models.Event{}).
		Where("id = ? AND available_seats > 0", eventID).
		Update("available_seats", gorm.Expr("available_seats - 1"))

	if result.Error != nil {
		tx.Rollback()
		return nil, result.Error
	}

	// Check if the update actually affected any rows
	// This is our final safety net - if no rows affected, seats are gone
	if result.RowsAffected == 0 {
		tx.Rollback()
		return nil, models.ErrEventFull
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// Reload the registration with associations
	s.registrationRepo.FindByUserAndEventID(userID, eventID)
	return registration, nil
}

// GetRegistrationByID gets a registration by ID
func (s *registrationService) GetRegistrationByID(id uint) (*models.Registration, error) {
	return s.registrationRepo.FindByID(id)
}

// GetUserRegistrations gets all registrations for a user
func (s *registrationService) GetUserRegistrations(userID uint) ([]models.Registration, error) {
	return s.registrationRepo.FindByUserID(userID)
}

// GetEventRegistrations gets all registrations for an event
func (s *registrationService) GetEventRegistrations(eventID uint) ([]models.Registration, error) {
	return s.registrationRepo.FindByEventID(eventID)
}

// CancelRegistration cancels a user's registration for an event
func (s *registrationService) CancelRegistration(userID, eventID uint) error {
	// Start transaction for atomic update
	tx := s.db.Begin()

	// Find and delete the registration
	err := tx.Where("user_id = ? AND event_id = ?", userID, eventID).Delete(&models.Registration{}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// Increment available seats
	result := tx.Model(&models.Event{}).
		Where("id = ?", eventID).
		Update("available_seats", gorm.Expr("available_seats + 1"))

	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	return tx.Commit().Error
}
