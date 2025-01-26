package order

import (
	"src/pkg/module/address"
	"src/pkg/module/cart"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Order struct {
	ID        uuid.UUID   `db:"id" json:"id"`
	CartID    uuid.UUID   `db:"cart_id" json:"cart_id"`
	UserID    uuid.UUID   `db:"user_id" json:"user_id"`
	AddressID uuid.UUID   `db:"address_id" json:"address_id"`
	Total     float64     `db:"total" json:"total"`
	Updated   pq.NullTime `db:"updated" json:"updated"`
	Created   time.Time   `db:"created" json:"created"`
}

type OrderItem struct {
	ID            uuid.UUID           `db:"id" json:"id"`
	OrderID       uuid.UUID           `db:"order_id" json:"order_id"`
	ProductID     uuid.UUID           `db:"product_id" json:"product_id"`
	Quantity      int                 `db:"quantity" json:"quantity"`
	PurchasePrice float64             `db:"purchase_price" json:"purchase_price"`
	Status        cart.CartItemStatus `db:"status" json:"status"` // Assuming CartItemStatus is defined similarly in cart2
	UpdatedAt     time.Time           `db:"updated" json:"updated"`
	CreatedAt     time.Time           `db:"created" json:"created"`
}

// type OrderRequest struct {
// 	CartID    uuid.UUID `json:"cart_id" binding:"required"`
// 	Total     float64   `json:"total" binding:"required"`
// 	AddressID uuid.UUID `json:"address_id" binding:"required"`
// }

// type AddOrderWithCartItemAndAddressRequest struct {
// 	CartID    uuid.UUID `json:"cart_id" binding:"required"`
// 	AddressID uuid.UUID `json:"address_id" binding:"required"`
// }

// type CartProduct struct {
// 	product.Product      // Embed the product struct
// 	CartItemQuantity int `db:"cart_item_quantity" json:"cart_item_quantity"` // Add the quantity from cart_items
// }

// type OrderGet struct {
// 	ID       uuid.UUID        `db:"id" json:"_id"`
// 	CartID   uuid.UUID        `db:"cart_id" json:"cart"`
// 	UserID   uuid.UUID        `db:"user_id" json:"user"`
// 	Total    float64          `db:"total" json:"total"`
// 	Updated  pq.NullTime      `db:"updated" json:"updated"`
// 	Created  time.Time        `db:"created" json:"created"`
// 	Address  address2.Address `json:"address"`  // Use your existing address struct
// 	Products []cart2.CartItem `json:"products"` // From cart2 package
// }

type OrderInfo struct { // Struct for fetching additional details
	ID       uuid.UUID       `json:"id"`
	CartID   uuid.UUID       `json:"cart_id"`
	UserID   uuid.UUID       `json:"user_id"`
	Total    float64         `json:"total"`
	Updated  pq.NullTime     `json:"updated"`
	Created  time.Time       `json:"created"`
	Address  address.Address `json:"address"`
	Products []cart.CartItem `json:"products"`
}

// // Request Structs

type AddOrderRequest struct {
	CartID    uuid.UUID `json:"cart_id" binding:"required"`
	Total     float64   `json:"total" binding:"required"`
	AddressID uuid.UUID `json:"address_id" binding:"required"`
}

type AddOrder2Request struct { // Request struct for AddOrderWithCartItemAndAddress
	CartID    uuid.UUID `json:"cart_id" binding:"required"`
	AddressID uuid.UUID `json:"address_id" binding:"required"`
}

type UpdateOrderItemStatusRequest struct {
	OrderID uuid.UUID           `json:"order_id" binding:"required"`
	CartID  uuid.UUID           `json:"cart_id" binding:"required"`
	Status  cart.CartItemStatus `json:"status"`
}
