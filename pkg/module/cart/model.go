package cart

import (
	"src/pkg/module/product"
	"src/pkg/module/user"
	"time"
)

// CartItemStatus type
type CartItemStatus string

// CartItemStatus constants
const (
	NotProcessed CartItemStatus = "Not processed"
	Processing   CartItemStatus = "Processing"
	Shipped      CartItemStatus = "Shipped"
	Delivered    CartItemStatus = "Delivered"
	Cancelled    CartItemStatus = "Cancelled"
)

// CartItem model
type CartItem struct {
	ID            uint            `json:"id"`
	CartID        uint            `json:"cartId"` // Add this line
	ProductID     uint            `json:"productId"`
	Product       product.Product `json:"product"`
	Quantity      int             `json:"quantity"`
	PurchasePrice float64         `json:"purchasePrice"`
	TotalPrice    float64         `json:"totalPrice"`
	PriceWithTax  float64         `json:"priceWithTax"`
	TotalTax      float64         `json:"totalTax"`
	Status        CartItemStatus  `json:"status"`
}

// Cart model
type Cart struct {
	ID      uint       `json:"id"`
	Items   []CartItem `json:"items"`
	UserID  uint       `json:"userId"`
	User    user.User  `json:"user"`
	Updated time.Time  `json:"updated"`
	Created time.Time  `json:"created"`
}
