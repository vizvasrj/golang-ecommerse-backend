package brand

import (
	"time"

	"github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Brand represents the brand model
type Brand struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Slug        string             `bson:"slug" json:"slug"`
	Image       Image              `bson:"image" json:"image"`
	Description string             `bson:"description" json:"description"`
	IsActive    bool               `bson:"isActive" json:"isActive"`
	Merchant    primitive.ObjectID `bson:"merchant,omitempty" json:"merchant,omitempty"`
	Updated     time.Time          `bson:"updated" json:"updated"`
	Created     time.Time          `bson:"created" json:"created"`
}

// Image represents the image model
type Image struct {
	Data        []byte `bson:"data" json:"data"`
	ContentType string `bson:"contentType" json:"contentType"`
}

func NewBrand(b Brand) *Brand {
	b.Slug = slug.Make(b.Name)
	return &b
}

type BrandUpdate struct {
	Name        string `bson:"name,omitempty" json:"name,omitempty"`
	Slug        string `bson:"slug,omitempty" json:"slug,omitempty"`
	Image       Image  `bson:"image,omitempty" json:"image,omitempty"`
	Description string `bson:"description,omitempty" json:"description,omitempty"`
	IsActive    bool   `bson:"isActive,omitempty" json:"isActive,omitempty"`
}
