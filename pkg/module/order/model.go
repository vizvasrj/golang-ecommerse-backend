package order

import (
	"time"
)

// Address model
type Address struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user"`
	Address   string    `json:"address"`
	City      string    `json:"city"`
	State     string    `json:"state"`
	Country   string    `json:"country"`
	ZipCode   string    `json:"zipCode"`
	IsDefault bool      `json:"isDefault"`
	Updated   time.Time `json:"updated"`
	Created   time.Time `json:"created"`
}

// Order model
type Order struct {
	ID      uint      `json:"id"`
	CartID  uint      `json:"cart"`
	UserID  uint      `json:"user"`
	Total   float64   `json:"total"`
	Updated time.Time `json:"updated"`
	Created time.Time `json:"created"`
	Address Address   `json:"address"`
}
