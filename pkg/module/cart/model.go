package cart

import (
	"src/pkg/module/product"
	"time"

	"github.com/google/uuid"
)

type CartItemStatus string

const (
	NotOrdered   CartItemStatus = "Not_ordered"
	NotProcessed CartItemStatus = "Not_processed"
	Processing   CartItemStatus = "Processing"
	Shipped      CartItemStatus = "Shipped"
	Delivered    CartItemStatus = "Delivered"
	Cancelled    CartItemStatus = "Cancelled"
)

type Cart struct {
	ID      uuid.UUID `db:"id" json:"_id"`
	UserID  uuid.UUID `db:"user_id" json:"userId"`
	Updated time.Time `db:"updated" json:"updated"`
	Created time.Time `db:"created" json:"created"`
}

type CartItem struct {
	ID            uuid.UUID        `db:"id" json:"_id"`
	CartID        uuid.UUID        `db:"cart_id" json:"cartId"`
	ProductID     uuid.UUID        `db:"product_id" json:"productId"`
	Quantity      int              `db:"quantity" json:"quantity"`
	PurchasePrice float64          `db:"purchase_price" json:"purchasePrice"`
	UpdatedAt     time.Time        `db:"updated"`
	CreatedAt     time.Time        `db:"created"`
	Product       *product.Product `json:"product,omitempty"`
	Status        CartItemStatus   `db:"status" json:"status"`
}

// Request Structs

type AddProductToCartRequest struct {
	ProductID uuid.UUID `json:"productId" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required"`
}

type CartItemRequest struct { // For AddProductToCart and RemoveProductFromCart functions
	ProductID uuid.UUID `json:"productId" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required"`
	Action    string    `json:"action"` // TODO "replace" or "increment"
}
