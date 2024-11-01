package category

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Category represents a category in the e-commerce system.
type Category struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	Name        string               `bson:"name,omitempty" json:"name,omitempty"`
	Slug        string               `bson:"slug,omitempty" json:"slug,omitempty"`
	Image       Image                `bson:"image,omitempty" json:"image,omitempty"`
	Description string               `bson:"description,omitempty" json:"description,omitempty"`
	IsActive    bool                 `bson:"isActive,omitempty" json:"isActive,omitempty"`
	Products    []primitive.ObjectID `bson:"products,omitempty" json:"products,omitempty"`
	Updated     time.Time            `bson:"updated,omitempty" json:"updated,omitempty"`
	Created     time.Time            `bson:"created,omitempty" json:"created,omitempty"`
}

// Image represents an image in the e-commerce system.
type Image struct {
	Data        []byte `json:"data,omitempty" bson:"data,omitempty"`
	ContentType string `json:"contentType,omitempty" bson:"contentType,omitempty"`
}
