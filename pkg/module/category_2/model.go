package category

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID          uuid.UUID `db:"id" json:"id"`
	Name        string    `db:"name" json:"name" binding:"required"`
	Slug        string    `db:"slug" json:"slug"`
	Description string    `db:"description" json:"description" binding:"required"`
	IsActive    bool      `db:"is_active" json:"is_active"`
	Updated     time.Time `db:"updated" json:"updated"`
	Created     time.Time `db:"created" json:"created"`
}

type CategoryUpdate struct { // Struct for partial updates
	Name        *string `json:"name"`
	Slug        *string `json:"slug"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"`
}
