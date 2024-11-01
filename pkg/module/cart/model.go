package cart

import (
	"src/pkg/module/product"

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
	Product       product.Product `bson:"product" json:"product"`
	Quantity      int             `bson:"quantity" json:"quantity"`
	PurchasePrice float64         `bson:"purchasePrice" json:"purchasePrice"`
	TotalPrice    float64         `bson:"totalPrice" json:"totalPrice"`
	PriceWithTax  float64         `bson:"priceWithTax" json:"priceWithTax"`
	TotalTax      float64         `bson:"totalTax" json:"totalTax"`
	Status        CartItemStatus  `bson:"status" json:"status"`
}

type Cart struct {
	Products []CartItem          `bson:"products" json:"products"`
	User     string              `bson:"user" json:"user"`
	Updated  primitive.Timestamp `bson:"updated" json:"updated"`
	Created  primitive.Timestamp `bson:"created" json:"created"`
}
