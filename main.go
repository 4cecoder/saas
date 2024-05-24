// Package main/main.go
package main

import (
	"gorm.io/gorm"
	"log"

	"github.com/4cecoder/saas/config"
	"github.com/4cecoder/saas/handlers"
	"github.com/4cecoder/saas/models"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Auto-migrate models
	err := cfg.DB.AutoMigrate(
		&models.User{},
		&models.Organization{},
		&models.Subscription{},
		&models.Role{},
		&models.Permission{},
		&models.Domain{},
		&models.AuditLog{},
		&models.PaymentTransaction{},
		&models.NotificationPreference{},
		&models.ActivityLog{},
		&models.APIKey{},
		&models.Workflow{},
		&models.Report{},
	)
	if err != nil {
		log.Fatalf("Failed to auto-migrate models: %v", err)
	}

	// Create a new Gin router
	r := gin.Default()

	// Create a new handler instance
	h := handlers.NewHandler(cfg.DB)

	// Define routes
	r.POST("/users", h.CreateUser)
	r.GET("/users/:id", h.GetUser)
	r.PUT("/users/:id", h.UpdateUser)
	r.DELETE("/users/:id", h.DeleteUser)

	r.POST("/organizations", h.CreateOrganization)
	r.GET("/organizations/:id", h.GetOrganization)
	r.PUT("/organizations/:id", h.UpdateOrganization)
	r.DELETE("/organizations/:id", h.DeleteOrganization)

	r.POST("/subscriptions", h.CreateSubscription)
	r.GET("/subscriptions/:id", h.GetSubscription)
	r.PUT("/subscriptions/:id", h.UpdateSubscription)
	r.DELETE("/subscriptions/:id", h.DeleteSubscription)

	// Add more routes for other handlers

	// Create the default admin user
	createDefaultAdmin(cfg.DB)

	// Start the server
	err = r.Run(":8080")
	if err != nil {
		log.Fatalf("Failed to start the server: %v", err)
	}
}

func createDefaultAdmin(db *gorm.DB) {
	// Check if the admin user already exists
	var count int64
	db.Model(&models.User{}).Where("email = ?", "admin").Count(&count)

	if count == 0 {
		// Create the admin user

		admin := &models.User{
			Email:        "admin",
			PasswordHash: "password",
			Name:         "Admin User",
			Roles: []models.Role{
				{Name: "admin"},
			},
			Verified: true,
		}

		if err := db.Create(admin).Error; err != nil {
			log.Fatalf("Failed to create admin user: %v", err)
		}

		log.Println("Default admin user created")
	}
}
