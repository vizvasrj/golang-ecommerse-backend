package address

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Address struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	User      primitive.ObjectID `bson:"user,omitempty" json:"user,omitempty"`
	Address   string             `bson:"address,omitempty" json:"address,omitempty"`
	City      string             `bson:"city,omitempty" json:"city,omitempty"`
	State     string             `bson:"state,omitempty" json:"state,omitempty"`
	Country   string             `bson:"country,omitempty" json:"country,omitempty"`
	ZipCode   string             `bson:"zipCode,omitempty" json:"zipCode,omitempty"`
	IsDefault bool               `bson:"isDefault,omitempty" json:"isDefault,omitempty"`
	Updated   time.Time          `bson:"updated,omitempty" json:"updated,omitempty"`
	Created   time.Time          `bson:"created,omitempty" json:"created,omitempty"`
}

type AddressUpdate struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	ZipCode string `json:"zip_code"`
}
type AddressAdd struct {
	Address   string `json:"address" bson:"address" binding:"required"`
	City      string `json:"city" bson:"city" binding:"required"`
	State     string `json:"state" bson:"state" binding:"required"`
	Country   string `json:"country" bson:"country" binding:"required"`
	ZipCode   string `json:"zipCode" bson:"zipCode" binding:"required"`
	IsDefault bool   `json:"isDefault" bson:"isDefault"`
}
