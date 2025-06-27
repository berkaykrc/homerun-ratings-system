package entity

import "time"

// ServiceProvider represents a service provider record.
type ServiceProvider struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ServiceProvider) TableName() string {
	return "service_providers"
}
