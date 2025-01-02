package brand2

import (
	"time"

	"github.com/google/uuid"
	"github.com/guregu/null/v5"
)

type Brand struct {
	ID          uuid.UUID   `db:"id" json:"id"`
	Name        string      `db:"name" json:"name" binding:"required"`
	Slug        string      `db:"slug" json:"slug"`
	Image       string      `db:"image" json:"image"` // Assuming image is a URL now
	ContentType null.String `db:"content_type" json:"content_type"`
	Description string      `db:"description" json:"description" binding:"required"`
	IsActive    bool        `db:"is_active" json:"is_active"`
	Updated     time.Time   `db:"updated" json:"updated"`
	Created     time.Time   `db:"created" json:"created"`
}

type BrandUpdate struct {
	Name        *string `db:"name" json:"name"` // Pointers for optional fields
	Slug        *string `db:"slug" json:"slug"`
	Image       *string `db:"image" json:"image"`
	ContentType *string `db:"content_type" json:"content_type"`
	Description *string `db:"description" json:"description"`
	IsActive    *bool   `db:"is_active" json:"is_active"`
}
