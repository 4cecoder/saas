// Package models/models.go
package models

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Base contains common fields for all models
type Base struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

// User represents a user in the system
type User struct {
	Base
	Email             string                 `gorm:"unique" json:"email"`
	Password          string                 `json:"-"`
	PasswordHash      string                 `json:"-"`
	Name              string                 `json:"name"`
	Roles             []Role                 `gorm:"many2many:user_roles;" json:"roles"`
	Organizations     []Organization         `gorm:"many2many:user_organizations;" json:"organizations"`
	Seats             []Seat                 `json:"seats"`
	Permissions       []Permission           `gorm:"many2many:user_permissions;" json:"permissions"`
	VerificationCode  string                 `json:"-"`
	Verified          bool                   `json:"verified"`
	ActivityLogs      []ActivityLog          `json:"activity_logs"`
	NotificationPrefs NotificationPreference `gorm:"foreignKey:UserID" json:"notification_prefs"`
	Locale            string                 `json:"locale"`
	Timezone          string                 `json:"timezone"`
	Language          string                 `json:"language"`
}

// BeforeCreate is a GORM hook that runs before creating a new user
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// Hash the password
	err := u.hashPassword()
	if err != nil {
		return err
	}

	// Generate verification code
	u.VerificationCode = generateRandomString(32)

	return nil
}

// BeforeUpdate is a GORM hook that runs before updating a user
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	// Hash the password if it's being updated
	if tx.Statement.Changed("Password") {
		err := u.hashPassword()
		if err != nil {
			return err
		}
	}

	return nil
}

// hashPassword hashes the user's password using bcrypt
func (u *User) hashPassword() error {
	if u.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.PasswordHash = string(hashedPassword)
		u.Password = ""
	}

	return nil
}

// Role defines the access level and permissions for a user
type Role struct {
	Base
	Name        string       `json:"name"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions"`
}

// Permission represents a specific action or resource access
type Permission struct {
	Base
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Organization represents a company or group
type Organization struct {
	Base
	Name             string               `json:"name"`
	Users            []User               `gorm:"many2many:user_organizations;" json:"users"`
	Subscriptions    []Subscription       `json:"subscriptions"`
	SubscriptionPlan SubscriptionPlan     `json:"subscription_plan"`
	Seats            []Seat               `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"seats"`
	Domains          []Domain             `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"domains"`
	Settings         OrganizationSettings `gorm:"embedded" json:"settings"`
	AuditLogs        []AuditLog           `json:"audit_logs"`
	ActivityLogs     []ActivityLog        `json:"activity_logs"`
	APIKeys          []APIKey             `json:"api_keys"`
	Workflows        []Workflow           `json:"workflows"`
}

// OrganizationSettings represents the settings for an organization
type OrganizationSettings struct {
	LogoURL    string `json:"logo_url"`
	ThemeColor string `json:"theme_color"`
	// Add more settings fields as needed
}

// Subscription represents a subscription for an organization
type Subscription struct {
	Base
	OrganizationID   uint                 `json:"organization_id"`
	SubscriptionPlan SubscriptionPlan     `json:"subscription_plan"`
	Status           SubscriptionStatus   `json:"status"`
	StartDate        time.Time            `json:"start_date"`
	EndDate          time.Time            `json:"end_date"`
	Transactions     []PaymentTransaction `json:"transactions"`
	PaymentMethod    string               `json:"payment_method"`
	LastPaymentDate  time.Time            `json:"last_payment_date"`
	NextBillingDate  time.Time            `json:"next_billing_date"`
}

// BeforeCreate is a GORM hook that runs before creating a new subscription
func (s *Subscription) BeforeCreate(tx *gorm.DB) error {
	// Set default status to trialing
	if s.Status == "" {
		s.Status = SubscriptionStatusTrialing
	}

	// Set start date to current time if not provided
	if s.StartDate.IsZero() {
		s.StartDate = time.Now()
	}

	return nil
}

// SubscriptionStatus represents the status of a subscription
type SubscriptionStatus string

const (
	SubscriptionStatusActive   SubscriptionStatus = "active"
	SubscriptionStatusInactive SubscriptionStatus = "inactive"
	SubscriptionStatusTrialing SubscriptionStatus = "trialing"
	SubscriptionStatusCanceled SubscriptionStatus = "canceled"
)

// SubscriptionPlan represents a subscription plan with pricing and features
type SubscriptionPlan struct {
	Base
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Currency    string    `json:"currency"`
	Interval    string    `json:"interval"`
	Features    []Feature `gorm:"many2many:subscription_plan_features;" json:"features"`
}

// Feature represents a specific feature of a subscription plan
type Feature struct {
	Base
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Seat represents a user seat within an organization
type Seat struct {
	Base
	OrganizationID uint       `json:"organization_id"`
	UserID         uint       `json:"user_id"`
	Roles          []Role     `gorm:"many2many:seat_roles;" json:"roles"`
	Status         SeatStatus `json:"status"`
}

// SeatStatus represents the status of a seat
type SeatStatus string

const (
	UserRole                      = "user"
	AdminRole                     = "admin"
	SeatStatusActive   SeatStatus = "active"
	SeatStatusInactive SeatStatus = "inactive"
	SeatStatusInvited  SeatStatus = "invited"
)

// Domain represents a custom domain for an organization
type Domain struct {
	Base
	OrganizationID uint   `json:"organization_id"`
	Domain         string `gorm:"unique" json:"domain"`
	Verified       bool   `json:"verified"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	Base
	UserID       uint         `json:"user_id"`
	Action       string       `json:"action"`
	ResourceType string       `json:"resource_type"`
	ResourceID   uint         `json:"resource_id"`
	Timestamp    time.Time    `json:"timestamp"`
	Changes      JSONMap      `json:"changes" gorm:"type:jsonb"`
	Organization Organization `gorm:"foreignKey:OrganizationID" json:"organization"`
}

// PaymentTransaction represents a payment transaction
type PaymentTransaction struct {
	Base
	SubscriptionID uint      `json:"subscription_id"`
	Amount         float64   `json:"amount"`
	Currency       string    `json:"currency"`
	Status         string    `json:"status"`
	Gateway        string    `json:"gateway"`
	GatewayID      string    `json:"gateway_id"`
	Timestamp      time.Time `json:"timestamp"`
}

// NotificationPreference represents user notification preferences
type NotificationPreference struct {
	Base
	UserID          uint `json:"user_id"`
	EmailEnabled    bool `json:"email_enabled"`
	SMSEnabled      bool `json:"sms_enabled"`
	InAppEnabled    bool `json:"in_app_enabled"`
	BillingEmails   bool `json:"billing_emails"`
	ProductEmails   bool `json:"product_emails"`
	MarketingEmails bool `json:"marketing_emails"`
}

// ActivityLog represents user activity log
type ActivityLog struct {
	Base
	UserID         uint      `json:"user_id"`
	OrganizationID uint      `json:"organization_id"`
	ActivityType   string    `json:"activity_type"`
	Timestamp      time.Time `json:"timestamp"`
	Metadata       JSONMap   `json:"metadata" gorm:"type:jsonb"`
}

// APIKey represents an API key for authentication
type APIKey struct {
	Base
	UserID         uint      `json:"user_id"`
	OrganizationID uint      `json:"organization_id"`
	Key            string    `gorm:"unique" json:"key"`
	Name           string    `json:"name"`
	Permissions    []string  `json:"permissions" gorm:"type:jsonb"`
	ExpiresAt      time.Time `json:"expires_at"`
	LastUsedAt     time.Time `json:"last_used_at"`
}

// BeforeCreate is a GORM hook that runs before creating a new API key
func (k *APIKey) BeforeCreate(tx *gorm.DB) error {
	// Generate a unique API key
	k.Key = generateRandomString(32)

	return nil
}

// Workflow represents a workflow process
type Workflow struct {
	Base
	Name           string         `json:"name"`
	Description    string         `json:"description"`
	Steps          []WorkflowStep `json:"steps" gorm:"type:jsonb"`
	OrganizationID uint           `json:"organization_id"`
	CreatorID      uint           `json:"creator_id"`
	Enabled        bool           `json:"enabled"`
}

// WorkflowStep represents a step in a workflow process
type WorkflowStep struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Order       int    `json:"order"`
	Approver    string `json:"approver"`
	Conditions  string `json:"conditions"`
}

// Report represents a report definition
type Report struct {
	Base
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Query          string    `json:"query"`
	OrganizationID uint      `json:"organization_id"`
	CreatorID      uint      `json:"creator_id"`
	Schedule       string    `json:"schedule"`
	Recipients     []string  `json:"recipients" gorm:"type:jsonb"`
	LastRunAt      time.Time `json:"last_run_at"`
}

// JSONMap is a type for storing JSON data in the database
type JSONMap map[string]interface{}

// generateRandomString generates a random string of the specified length
func generateRandomString(length int) string {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(bytes)
}
