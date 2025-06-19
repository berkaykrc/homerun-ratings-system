package entity

import "time"

// Rating represents an rating record.
type Rating struct {
	ID                string    `json:"id"`
	CustomerID        string    `json:"customerId"`
	ServiceProviderID string    `json:"serviceProviderId"`
	RatingValue       int       `json:"rating"`
	Comment           string    `json:"comment"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// TableName returns the table name for the Rating entity.
func (Rating) TableName() string {
	return "ratings"
}
