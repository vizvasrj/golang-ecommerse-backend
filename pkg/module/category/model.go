package category

import (
	"time"
)

// Category model
type Category struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	ImageData   []byte    `json:"imageData"`
	ImageType   string    `json:"imageType"`
	Description string    `json:"description"`
	IsActive    bool      `json:"isActive"`
	Products    []Product `json:"products"`
	Updated     time.Time `json:"updated"`
	Created     time.Time `json:"created"`
}
