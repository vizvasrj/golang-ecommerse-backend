package cart

import (
	"src/pkg/module/product"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CartItemStatus string

const (
	NotProcessed CartItemStatus = "Not_processed"
	Processing   CartItemStatus = "Processing"
	Shipped      CartItemStatus = "Shipped"
	Delivered    CartItemStatus = "Delivered"
	Cancelled    CartItemStatus = "Cancelled"
)

type CartItem struct {
	Product       product.IndividualProduct `bson:"product" json:"product"`
	Quantity      int                       `bson:"quantity" json:"quantity"`
	PurchasePrice float64                   `bson:"purchasePrice" json:"purchasePrice"`
	TotalPrice    float64                   `bson:"totalPrice" json:"totalPrice"`
	PriceWithTax  float64                   `bson:"priceWithTax" json:"priceWithTax"`
	TotalTax      float64                   `bson:"totalTax" json:"totalTax"`
	Status        CartItemStatus            `bson:"status" json:"status"`
}

type Cart struct {
	ID       primitive.ObjectID `bson:"_id" json:"id"`
	Products []CartItem         `bson:"products" json:"products"`
	User     primitive.ObjectID `bson:"user" json:"user"`
	Updated  time.Time          `bson:"updated" json:"updated"`
	Created  time.Time          `bson:"created" json:"created"`
}

type GetCart struct {
	ID       primitive.ObjectID `bson:"_id" json:"id"`
	Products []GetCartItem      `bson:"products" json:"products"`
	User     primitive.ObjectID `bson:"user" json:"user"`
	Updated  time.Time          `bson:"updated" json:"updated"`
	Created  time.Time          `bson:"created" json:"created"`
}

type GetCartItem struct {
	Product       primitive.ObjectID `bson:"product" json:"product"`
	Quantity      int                `bson:"quantity" json:"quantity"`
	PurchasePrice float64            `bson:"purchasePrice" json:"purchasePrice"`
	TotalPrice    float64            `bson:"totalPrice" json:"totalPrice"`
	PriceWithTax  float64            `bson:"priceWithTax" json:"priceWithTax"`
	TotalTax      float64            `bson:"totalTax" json:"totalTax"`
	Status        CartItemStatus     `bson:"status" json:"status"`
}

type AddProductToCartRequest struct {
	Product  primitive.ObjectID `json:"product" bson:"product" binding:"required"`
	Quantity int                `json:"quantity" bson:"quantity" binding:"required"`
}
