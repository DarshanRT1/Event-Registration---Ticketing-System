package repository

import (
	"event-api/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// EventRepository defines the interface for event data access
type EventRepository interface {
	Create(event *models.Event) error
	FindByID(id uint) (*models.Event, error)
	FindAll() ([]models.Event, error)
	FindByOrganizerID(organizerID uint) ([]models.Event, error)
	Update(event *models.Event) error
	Delete(id uint) error

	// Transaction-based operations for concurrency control
	FindByIDForUpdate(tx *gorm.DB, id uint) (*models.Event, error)
	DecreaseAvailableSeats(tx *gorm.DB, id uint) error
}

// eventRepository implements EventRepository
type eventRepository struct {
	db *gorm.DB
}

// NewEventRepository creates a new EventRepository
func NewEventRepository(db *gorm.DB) EventRepository {
	return &eventRepository{db: db}
}

// Create creates a new event
func (r *eventRepository) Create(event *models.Event) error {
	return r.db.Create(event).Error
}

// FindByID finds an event by ID
func (r *eventRepository) FindByID(id uint) (*models.Event, error) {
	var event models.Event
	err := r.db.Preload("Organizer").First(&event, id).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// FindAll returns all events
func (r *eventRepository) FindAll() ([]models.Event, error) {
	var events []models.Event
	err := r.db.Preload("Organizer").Find(&events).Error
	return events, err
}

// FindByOrganizerID returns all events created by an organizer
func (r *eventRepository) FindByOrganizerID(organizerID uint) ([]models.Event, error) {
	var events []models.Event
	err := r.db.Where("organizer_id = ?", organizerID).Find(&events).Error
	return events, err
}

// Update updates an event
func (r *eventRepository) Update(event *models.Event) error {
	return r.db.Save(event).Error
}

// Delete deletes an event by ID
func (r *eventRepository) Delete(id uint) error {
	return r.db.Delete(&models.Event{}, id).Error
}

// FindByIDForUpdate finds an event by ID with a row lock for updates
// This is critical for concurrency control - it uses SELECT FOR UPDATE
// to lock the row and prevent race conditions
func (r *eventRepository) FindByIDForUpdate(tx *gorm.DB, id uint) (*models.Event, error) {
	var event models.Event
	// ForUpdate() generates SELECT ... FOR UPDATE clause
	// This locks the row until the transaction is committed or rolled back
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&event, id).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// DecreaseAvailableSeats atomically decreases the available seats count
// This is done within a transaction to ensure consistency
func (r *eventRepository) DecreaseAvailableSeats(tx *gorm.DB, id uint) error {
	// Use UPDATE with a WHERE clause to ensure we only decrement if seats > 0
	// This provides an additional layer of safety against overbooking
	result := tx.Model(&models.Event{}).
		Where("id = ? AND available_seats > 0", id).
		Update("available_seats", gorm.Expr("available_seats - 1"))

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return models.ErrEventFull
	}

	return nil
}
