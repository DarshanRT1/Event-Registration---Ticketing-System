package repository

import (
	"event-api/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// RegistrationRepository defines the interface for registration data access
type RegistrationRepository interface {
	Create(registration *models.Registration) error
	FindByID(id uint) (*models.Registration, error)
	FindByUserID(userID uint) ([]models.Registration, error)
	FindByEventID(eventID uint) ([]models.Registration, error)
	FindByUserAndEventID(userID, eventID uint) (*models.Registration, error)
	Delete(id uint) error
	DeleteByUserAndEvent(userID, eventID uint) error
	
	// Transaction support
	CreateWithTx(tx *gorm.DB, registration *models.Registration) error
}

// registrationRepository implements RegistrationRepository
type registrationRepository struct {
	db *gorm.DB
}

// NewRegistrationRepository creates a new RegistrationRepository
func NewRegistrationRepository(db *gorm.DB) RegistrationRepository {
	return &registrationRepository{db: db}
}

// Create creates a new registration
func (r *registrationRepository) Create(registration *models.Registration) error {
	return r.db.Create(registration).Error
}

// FindByID finds a registration by ID
func (r *registrationRepository) FindByID(id uint) (*models.Registration, error) {
	var registration models.Registration
	err := r.db.Preload("User").Preload("Event").First(&registration, id).Error
	if err != nil {
		return nil, err
	}
	return &registration, nil
}

// FindByUserID returns all registrations for a user
func (r *registrationRepository) FindByUserID(userID uint) ([]models.Registration, error) {
	var registrations []models.Registration
	err := r.db.Preload("Event").Where("user_id = ?", userID).Find(&registrations).Error
	return registrations, err
}

// FindByEventID returns all registrations for an event
func (r *registrationRepository) FindByEventID(eventID uint) ([]models.Registration, error) {
	var registrations []models.Registration
	err := r.db.Preload("User").Where("event_id = ?", eventID).Find(&registrations).Error
	return registrations, err
}

// FindByUserAndEventID finds a registration by user and event ID
func (r *registrationRepository) FindByUserAndEventID(userID, eventID uint) (*models.Registration, error) {
	var registration models.Registration
	err := r.db.Where("user_id = ? AND event_id = ?", userID, eventID).First(&registration).Error
	if err != nil {
		return nil, err
	}
	return &registration, nil
}

// Delete deletes a registration by ID
func (r *registrationRepository) Delete(id uint) error {
	return r.db.Delete(&models.Registration{}, id).Error
}

// DeleteByUserAndEvent deletes a registration by user and event ID
func (r *registrationRepository) DeleteByUserAndEvent(userID, eventID uint) error {
	return r.db.Where("user_id = ? AND event_id = ?", userID, eventID).Delete(&models.Registration{}).Error
}

// CreateWithTx creates a new registration within a transaction
// This is the critical method for atomic registration with seat decrement
func (r *registrationRepository) CreateWithTx(tx *gorm.DB, registration *models.Registration) error {
	// Use ON CONFLICT DO NOTHING to handle race conditions on unique constraint
	// The actual seat availability check happens in the service layer
	return tx.Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(registration).Error
}
