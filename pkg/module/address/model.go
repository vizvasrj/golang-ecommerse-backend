package address

import (
	"src/pkg/module/user"
	"time"
)

// Address model
type Address struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"userId"`
	User      user.User `json:"user"`
	Address   string    `json:"address"`
	City      string    `json:"city"`
	State     string    `json:"state"`
	Country   string    `json:"country"`
	ZipCode   string    `json:"zipCode"`
	IsDefault bool      `json:"isDefault"`
	Updated   time.Time `json:"updated"`
	Created   time.Time `json:"created"`
}
