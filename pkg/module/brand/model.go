package brand

import (
	"time"

	"github.com/google/uuid"
	"github.com/guregu/null/v5"
)

type Brand struct {
	ID          uuid.UUID   `db:"id" json:"id,omitempty"`
	Name        string      `db:"name" json:"name,omitempty" binding:"required"`
	Slug        string      `db:"slug" json:"slug,omitempty"`
	Image       string      `db:"image" json:"image,omitempty"` // Assuming image is a URL now
	ContentType null.String `db:"content_type" json:"content_type,omitempty"`
	Description string      `db:"description" json:"description,omitempty" binding:"required"`
	IsActive    bool        `db:"is_active" json:"is_active,omitempty"`
	Updated     time.Time   `db:"updated" json:"updated,omitempty"`
	Created     time.Time   `db:"created" json:"created,omitempty"`
}

type BrandUpdate struct {
	Name        *string `db:"name" json:"name"` // Pointers for optional fields
	Slug        *string `db:"slug" json:"slug"`
	Image       *string `db:"image" json:"image"`
	ContentType *string `db:"content_type" json:"content_type"`
	Description *string `db:"description" json:"description"`
	IsActive    *bool   `db:"is_active" json:"is_active"`
}
