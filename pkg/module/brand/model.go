package brand

import (
	"time"
)

// Brand model
type Brand struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	ImageData   []byte    `json:"-"`
	ContentType string    `json:"-"`
	Description string    `json:"description"`
	IsActive    bool      `json:"isActive"`
	MerchantID  uint      `json:"merchant"`
	Updated     time.Time `json:"updated"`
	Created     time.Time `json:"created"`
}
