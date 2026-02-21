package service

import (
	"event-api/models"
	"event-api/repository"
)

// EventService handles event business logic
type EventService interface {
	CreateEvent(event *models.Event) error
	GetEventByID(id uint) (*models.Event, error)
	GetAllEvents() ([]models.Event, error)
	GetEventsByOrganizerID(organizerID uint) ([]models.Event, error)
	UpdateEvent(event *models.Event) error
	DeleteEvent(id uint) error
}

type eventService struct {
	eventRepo repository.EventRepository
}

// NewEventService creates a new EventService
func NewEventService(eventRepo repository.EventRepository) EventService {
	return &eventService{eventRepo: eventRepo}
}

// CreateEvent creates a new event
func (s *eventService) CreateEvent(event *models.Event) error {
	// Set available seats equal to capacity on creation
	event.AvailableSeats = event.Capacity
	return s.eventRepo.Create(event)
}

// GetEventByID gets an event by ID
func (s *eventService) GetEventByID(id uint) (*models.Event, error) {
	return s.eventRepo.FindByID(id)
}

// GetAllEvents gets all events
func (s *eventService) GetAllEvents() ([]models.Event, error) {
	return s.eventRepo.FindAll()
}

// GetEventsByOrganizerID gets events by organizer ID
func (s *eventService) GetEventsByOrganizerID(organizerID uint) ([]models.Event, error) {
	return s.eventRepo.FindByOrganizerID(organizerID)
}

// UpdateEvent updates an event
func (s *eventService) UpdateEvent(event *models.Event) error {
	return s.eventRepo.Update(event)
}

// DeleteEvent deletes an event
func (s *eventService) DeleteEvent(id uint) error {
	return s.eventRepo.Delete(id)
}
