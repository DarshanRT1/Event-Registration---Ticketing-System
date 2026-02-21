package main

import (
	"fmt"
	"log"

	"event-api/config"
	"event-api/handler"
	"event-api/repository"
	"event-api/service"

	"github.com/gin-gonic/gin"
)

func main() {
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

	// Initialize handlers
	userHandler := handler.NewUserHandler(userService)
	eventHandler := handler.NewEventHandler(eventService)
	registrationHandler := handler.NewRegistrationHandler(registrationService)

	// Setup router
	router := setupRouter(userHandler, eventHandler, registrationHandler)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.ServerPort)
	log.Printf("Server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// setupRouter configures all routes
func setupRouter(
	userHandler *handler.UserHandler,
	eventHandler *handler.EventHandler,
	registrationHandler *handler.RegistrationHandler,
) *gin.Engine {
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API info endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Event Registration API",
			"version": "1.0",
			"endpoints": map[string]string{
				"users":         "/api/v1/users",
				"events":        "/api/v1/events",
				"registrations": "/api/v1/registrations",
				"health":        "/health",
			},
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// User routes
		users := v1.Group("/users")
		{
			users.POST("", userHandler.CreateUser)
			users.GET("", userHandler.GetAllUsers)
			users.GET("/:id", userHandler.GetUser)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
		}

		// Event routes
		events := v1.Group("/events")
		{
			events.POST("", eventHandler.CreateEvent)
			events.GET("", eventHandler.GetAllEvents)
			events.GET("/:id", eventHandler.GetEvent)
			events.PUT("/:id", eventHandler.UpdateEvent)
			events.DELETE("/:id", eventHandler.DeleteEvent)
			events.GET("/organizer/:organizerID", eventHandler.GetOrganizerEvents)
		}

		// Registration routes
		registrations := v1.Group("/registrations")
		{
			registrations.POST("", registrationHandler.RegisterForEvent)
			registrations.GET("/:id", registrationHandler.GetRegistration)
			registrations.GET("/user/:userID", registrationHandler.GetUserRegistrations)
			registrations.GET("/event/:eventID", registrationHandler.GetEventRegistrations)
			registrations.DELETE("", registrationHandler.CancelRegistration)
		}
	}

	return router
}
