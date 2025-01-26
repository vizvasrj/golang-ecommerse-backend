package cart

import (
	"src/pkg/module/product"
	"time"

	"github.com/google/uuid"
)

type CartItemStatus string

const (
	NotProcessed CartItemStatus = "Not_processed"
	Processing   CartItemStatus = "Processing"
	Shipped      CartItemStatus = "Shipped"
	Delivered    CartItemStatus = "Delivered"
	Cancelled    CartItemStatus = "Cancelled"
)

type Cart struct {
	ID      uuid.UUID `db:"id" json:"id"`
	UserID  uuid.UUID `db:"user_id" json:"user_id"`
	Updated time.Time `db:"updated" json:"updated"`
	Created time.Time `db:"created" json:"created"`
}

type CartItem struct {
	ID            uuid.UUID        `db:"id" json:"id"`
	CartID        uuid.UUID        `db:"cart_id" json:"cart_id"`
	ProductID     uuid.UUID        `db:"product_id" json:"product_id"`
	Quantity      int              `db:"quantity" json:"quantity"`
	PurchasePrice float64          `db:"purchase_price" json:"purchase_price"`
	UpdatedAt     time.Time        `db:"updated"`
	CreatedAt     time.Time        `db:"created"`
	Product       *product.Product `json:"product,omitempty"`
	Status        CartItemStatus   `db:"status" json:"status"`
}

// Request Structs

type AddProductToCartRequest struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required"`
}

type CartItemRequest struct { // For AddProductToCart and RemoveProductFromCart functions
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required"`
	Action    string    `json:"action"` // TODO "replace" or "increment"
}
