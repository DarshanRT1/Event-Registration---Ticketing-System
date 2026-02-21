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

// RegistrationHandler handles HTTP requests for registrations
type RegistrationHandler struct {
	registrationService service.RegistrationService
}

// NewRegistrationHandler creates a new RegistrationHandler
func NewRegistrationHandler(registrationService service.RegistrationService) *RegistrationHandler {
	return &RegistrationHandler{registrationService: registrationService}
}

// RegisterForEvent handles POST /registrations
type RegisterRequest struct {
	UserID  uint `json:"user_id" binding:"required"`
	EventID uint `json:"event_id" binding:"required"`
}

// RegisterForEvent registers a user for an event
func (h *RegistrationHandler) RegisterForEvent(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	registration, err := h.registrationService.RegisterForEvent(req.UserID, req.EventID)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, models.ErrEventNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, models.ErrEventFull):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, models.ErrAlreadyRegistered):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, registration)
}

// GetRegistration handles GET /registrations/:id
func (h *RegistrationHandler) GetRegistration(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid registration ID"})
		return
	}

	registration, err := h.registrationService.GetRegistrationByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "registration not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, registration)
}

// GetUserRegistrations handles GET /registrations/user/:userID
func (h *RegistrationHandler) GetUserRegistrations(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	registrations, err := h.registrationService.GetUserRegistrations(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, registrations)
}

// GetEventRegistrations handles GET /registrations/event/:eventID
func (h *RegistrationHandler) GetEventRegistrations(c *gin.Context) {
	eventID, err := strconv.ParseUint(c.Param("eventID"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event ID"})
		return
	}

	registrations, err := h.registrationService.GetEventRegistrations(uint(eventID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, registrations)
}

// CancelRegistration handles DELETE /registrations
type CancelRequest struct {
	UserID  uint `json:"user_id" binding:"required"`
	EventID uint `json:"event_id" binding:"required"`
}

// CancelRegistration cancels a user's registration for an event
func (h *RegistrationHandler) CancelRegistration(c *gin.Context) {
	var req CancelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.registrationService.CancelRegistration(req.UserID, req.EventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "registration cancelled successfully"})
}
