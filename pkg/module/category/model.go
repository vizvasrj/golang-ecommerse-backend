package category

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID          uuid.UUID `db:"id" json:"id"`
	Name        string    `db:"name" json:"name" binding:"required"`
	Slug        string    `db:"slug" json:"slug"`
	Description string    `db:"description" json:"description,omitempty"`
	IsActive    bool      `db:"is_active" json:"is_active,omitempty"`
	Updated     time.Time `db:"updated" json:"updated,omitempty"`
	Created     time.Time `db:"created" json:"created,omitempty"`
}

type CategoryUpdate struct { // Struct for partial updates
	Name        *string `json:"name"`
	Slug        *string `json:"slug"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"`
}
