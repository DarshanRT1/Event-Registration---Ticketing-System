package service

import (
	"event-api/models"
	"event-api/repository"
)

// UserService handles user business logic
type UserService interface {
	CreateUser(user *models.User) error
	GetUserByID(id uint) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetAllUsers() ([]models.User, error)
	UpdateUser(user *models.User) error
	DeleteUser(id uint) error
}

type userService struct {
	userRepo repository.UserRepository
}

// NewUserService creates a new UserService
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

// CreateUser creates a new user
func (s *userService) CreateUser(user *models.User) error {
	return s.userRepo.Create(user)
}

// GetUserByID gets a user by ID
func (s *userService) GetUserByID(id uint) (*models.User, error) {
	return s.userRepo.FindByID(id)
}

// GetUserByEmail gets a user by email
func (s *userService) GetUserByEmail(email string) (*models.User, error) {
	return s.userRepo.FindByEmail(email)
}

// GetAllUsers gets all users
func (s *userService) GetAllUsers() ([]models.User, error) {
	return s.userRepo.FindAll()
}

// UpdateUser updates a user
func (s *userService) UpdateUser(user *models.User) error {
	return s.userRepo.Update(user)
}

// DeleteUser deletes a user
func (s *userService) DeleteUser(id uint) error {
	return s.userRepo.Delete(id)
}
