package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Custom errors for registration
var (
	ErrAlreadyRegistered = errors.New("user already registered for this event")
	ErrEventFull        = errors.New("event is full")
	ErrEventNotFound    = errors.New("event not found")
	ErrUserNotFound     = errors.New("user not found")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrInvalidInput     = errors.New("invalid input")
)

// UserRole represents the role of a user in the system
type UserRole string

const (
	RoleOrganizer UserRole = "organizer"
	RoleAttendee  UserRole = "attendee"
)

// User represents a user in the event registration system
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"type:varchar(255);not null" json:"name"`
	Email     string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Role      UserRole       `gorm:"type:varchar(50);not null;default:'attendee'" json:"role"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Events    []Event        `gorm:"foreignKey:OrganizerID" json:"-"`
}

// Event represents an event in the ticketing system
type Event struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	Title          string         `gorm:"type:varchar(255);not null" json:"title"`
	Capacity       int            `gorm:"not null" json:"capacity"`
	AvailableSeats int            `gorm:"not null" json:"available_seats"`
	OrganizerID    uint           `gorm:"not null" json:"organizer_id"`
	Organizer      *User          `gorm:"foreignKey:OrganizerID" json:"organizer,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
	Registrations  []Registration `gorm:"foreignKey:EventID" json:"-"`
}

// Registration represents a user's registration for an event
type Registration struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null" json:"user_id"`
	EventID   uint           `gorm:"not null" json:"event_id"`
	User      *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Event     *Event         `gorm:"foreignKey:EventID" json:"event,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Unique constraint on (user_id, event_id) - handled via GORM constraints
}

// TableName specifies the table name for Registration
func (Registration) TableName() string {
	return "registrations"
}
