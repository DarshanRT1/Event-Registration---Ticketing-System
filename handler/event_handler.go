package handler

import (
	"errors"
	"net/http"
	"strconv"

	"event-api/models"
	"event-api/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// EventHandler handles HTTP requests for events
type EventHandler struct {
	eventService service.EventService
}

// NewEventHandler creates a new EventHandler
func NewEventHandler(eventService service.EventService) *EventHandler {
	return &EventHandler{eventService: eventService}
}

// CreateEvent handles POST /events
func (h *EventHandler) CreateEvent(c *gin.Context) {
	var event models.Event
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate capacity
	if event.Capacity <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "capacity must be greater than 0"})
		return
	}

	// Set available seats equal to capacity
	event.AvailableSeats = event.Capacity

	if err := h.eventService.CreateEvent(&event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, event)
}

// GetEvent handles GET /events/:id
func (h *EventHandler) GetEvent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event ID"})
		return
	}

	event, err := h.eventService.GetEventByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, event)
}

// GetAllEvents handles GET /events
func (h *EventHandler) GetAllEvents(c *gin.Context) {
	events, err := h.eventService.GetAllEvents()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, events)
}

// GetOrganizerEvents handles GET /events/organizer/:organizerID
func (h *EventHandler) GetOrganizerEvents(c *gin.Context) {
	organizerID, err := strconv.ParseUint(c.Param("organizerID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organizer ID"})
		return
	}

	events, err := h.eventService.GetEventsByOrganizerID(uint(organizerID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, events)
}

// UpdateEvent handles PUT /events/:id
func (h *EventHandler) UpdateEvent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event ID"})
		return
	}

	var event models.Event
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event.ID = uint(id)

	// Don't allow updating capacity to less than current registrations
	existingEvent, err := h.eventService.GetEventByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}

	// If capacity is being reduced, check that it doesn't go below current registrations
	if event.Capacity < existingEvent.Capacity {
		registrationsCount := existingEvent.Capacity - existingEvent.AvailableSeats
		if event.Capacity < registrationsCount {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot reduce capacity below current registrations"})
			return
		}
		// Recalculate available seats
		event.AvailableSeats = event.Capacity - registrationsCount
	}

	if err := h.eventService.UpdateEvent(&event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, event)
}

// DeleteEvent handles DELETE /events/:id
func (h *EventHandler) DeleteEvent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event ID"})
		return
	}

	if err := h.eventService.DeleteEvent(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "event deleted successfully"})
}
