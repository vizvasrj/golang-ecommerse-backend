package order

import (
	"src/pkg/module/address"
	"src/pkg/module/cart"
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

type Order struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Cart     primitive.ObjectID `bson:"cart,omitempty" json:"cart,omitempty"`
	User     primitive.ObjectID `bson:"user,omitempty" json:"user,omitempty"`
	Total    float64            `bson:"total,omitempty" json:"total,omitempty"`
	Updated  time.Time          `bson:"updated,omitempty" json:"updated,omitempty"`
	Created  time.Time          `bson:"created,omitempty" json:"created,omitempty"`
	Address  Address            `bson:"address,omitempty" json:"address,omitempty"`
	Products []cart.GetCartItem `bson:"products,omitempty" json:"products,omitempty"`
}

type OrderAdd struct {
	UserID  primitive.ObjectID `json:"userId"`
	CartID  string             `json:"cartId"`
	Total   float64            `json:"total"`
	Address Address            `json:"address"`
}

type OrderGet struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Cart     primitive.ObjectID `bson:"cart,omitempty" json:"cart,omitempty"`
	User     primitive.ObjectID `bson:"user,omitempty" json:"user,omitempty"`
	Total    float64            `bson:"total,omitempty" json:"total,omitempty"`
	Updated  time.Time          `bson:"updated,omitempty" json:"updated,omitempty"`
	Created  time.Time          `bson:"created,omitempty" json:"created,omitempty"`
	Address  Address            `bson:"address,omitempty" json:"address,omitempty"`
	Products []cart.CartItem    `bson:"products,omitempty" json:"products,omitempty"`
}

// uses in order add new api
type newOrder struct {
	UserID  primitive.ObjectID `json:"userId"`
	CartID  primitive.ObjectID `json:"cartId"`
	Total   float64            `json:"total"`
	Address address.Address    `json:"address"`
	Created time.Time          `json:"created"`
}
