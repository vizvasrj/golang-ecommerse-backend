package contact

import (
	"time"

	"github.com/google/uuid"
)

// Contact represents the contact model
type Contact struct {
	ID      uuid.UUID `bson:"_id,omitempty" json:"_id,omitempty"`
	Name    string    `bson:"name,omitempty" json:"name,omitempty"`
	Email   string    `bson:"email,omitempty" json:"email,omitempty"`
	Message string    `bson:"message,omitempty" json:"message,omitempty"`
	Updated time.Time `bson:"updated,omitempty" json:"updated,omitempty"`
	Created time.Time `bson:"created,omitempty" json:"created,omitempty"`
}
