package order

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"src/common"
	"src/l"
	"src/pkg/conf"
	"src/pkg/module/cart"
	"src/pkg/module/payment"
	"src/pkg/module/product"
)

// Model Structs

// HTTP Handlers

func AddOrder(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req AddOrderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		userIDStr := c.GetString("userID")
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		ctx := context.Background()
		tx, err := app.DB.BeginTx(ctx, nil)
		if err != nil {
			l.ErrorF("Transaction begin error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
			return
		}
		defer tx.Rollback() // Defer rollback

		newOrderID := uuid.New()

		_, err = tx.ExecContext(ctx, `
			INSERT INTO orders (id, cart_id, user_id, address_id, total, created)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, newOrderID, req.CartID, userID, req.AddressID, req.Total, time.Now())
		if err != nil {
			l.ErrorF("Failed to create order: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to place order"})
			return
		}

		if err := tx.Commit(); err != nil { // Commit transaction
			l.ErrorF("Transaction commit error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Order placed successfully!", "order_id": newOrderID})
	}
}

func AddOrderWithCartItemAndAddress(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req AddOrder2Request

		if !bindAndValidateRequest(c, &req) {
			return
		}

		userID, err := getUserID(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		ctx := context.Background()
		tx, err := beginTransaction(ctx, app)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
			return
		}
		defer tx.Rollback()

		cartItems, err := fetchCartItems(ctx, tx, req.CartID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve cart items"})
			return
		}

		total := calculateTotal(cartItems)

		if !verifyAddressOwnership(ctx, tx, req.Address.ID, userID) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "You are not authorized to use this address."})
			return
		}

		if !verifyCartOwnership(ctx, tx, req.CartID, userID) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "You are not authorized to use this cart."})
			return
		}

		newOrderID, err := createOrder(ctx, tx, req, userID, total)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
			return
		}

		razorpayOrderID, providerData, err := initiatePayment(total, newOrderID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate payment"})
			return
		}

		if err := createReceipt(ctx, tx, newOrderID, total, providerData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create receipt"})
			return
		}

		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}

		respondSuccess(c, newOrderID, total, razorpayOrderID)
	}
}

func bindAndValidateRequest(c *gin.Context, req *AddOrder2Request) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return false
	}
	return true
}

func getUserID(c *gin.Context) (uuid.UUID, error) {
	userIDStr := c.GetString("userID")
	return uuid.Parse(userIDStr)
}

func beginTransaction(ctx context.Context, app *conf.Config) (*sql.Tx, error) {
	return app.DB.BeginTx(ctx, nil)
}

func fetchCartItems(ctx context.Context, tx *sql.Tx, cartID uuid.UUID) ([]cart.CartItem, error) {
	rows, err := tx.QueryContext(ctx, `
        SELECT ci.product_id, ci.quantity, p.price
        FROM cart_items ci
        JOIN products p ON ci.product_id = p.id
        WHERE ci.cart_id = $1
    `, cartID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cartItems []cart.CartItem
	for rows.Next() {
		var cartItem cart.CartItem
		var price float64
		if err := rows.Scan(&cartItem.ProductID, &cartItem.Quantity, &price); err != nil {
			return nil, err
		}
		cartItem.PurchasePrice = price
		cartItems = append(cartItems, cartItem)
	}
	return cartItems, nil
}

func calculateTotal(cartItems []cart.CartItem) float64 {
	total := 0.0
	for _, item := range cartItems {
		total += item.PurchasePrice * float64(item.Quantity)
	}
	return total
}

func verifyAddressOwnership(ctx context.Context, tx *sql.Tx, addressID, userID uuid.UUID) bool {
	var addressExists bool
	err := tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM addresses WHERE id = $1 AND user_id = $2)", addressID, userID).Scan(&addressExists)
	return err == nil && addressExists
}

func verifyCartOwnership(ctx context.Context, tx *sql.Tx, cartID, userID uuid.UUID) bool {
	var cartExists bool
	err := tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM carts WHERE id = $1 AND user_id = $2)", cartID, userID).Scan(&cartExists)
	return err == nil && cartExists
}

func createOrder(ctx context.Context, tx *sql.Tx, req AddOrder2Request, userID uuid.UUID, total float64) (uuid.UUID, error) {
	newOrderID := uuid.New()
	_, err := tx.ExecContext(ctx, `
        INSERT INTO orders (id, cart_id, user_id, address_id, total, created)
        VALUES ($1, $2, $3, $4, $5, $6)
    `, newOrderID, req.CartID, userID, req.Address.ID, total, time.Now())
	return newOrderID, err
}

func initiatePayment(total float64, orderID uuid.UUID) (string, interface{}, error) {
	newReceiptID := uuid.New()
	return payment.Executerazorpay(total, newReceiptID, orderID.String())
}

func createReceipt(ctx context.Context, tx *sql.Tx, orderID uuid.UUID, total float64, providerData interface{}) error {
	providerDataJSON, err := json.Marshal(providerData)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
        INSERT INTO receipts (id, order_id, amount, created, updated, payment_provider, provider_data, payment_status)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `, uuid.New(), orderID, total, time.Now(), time.Now(), "razorpay", string(providerDataJSON), "PENDING")
	return err
}

func respondSuccess(c *gin.Context, orderID uuid.UUID, total float64, razorpayOrderID string) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Your order has been placed successfully!",
		"order": gin.H{
			"_id":    orderID,
			"amount": total * 100,
		},
		"razorpay_order_id": razorpayOrderID,
		"razorpay_id":       os.Getenv("RAZORPAY_ID"),
	})
}
func SearchOrders(app *conf.Config) gin.HandlerFunc { // Updated SearchOrders function
	return func(c *gin.Context) {
		searchQuery := c.Query("search")
		orderID, err := uuid.Parse(searchQuery)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID format"})
			return
		}

		userIDStr := c.GetString("userID")
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		userRoleStr := c.MustGet("role").(string)

		userRole := common.GetUserRole(userRoleStr)

		query := "SELECT * FROM orders WHERE id = $1" // Start with the most specific filter

		var args []interface{}
		args = append(args, orderID)

		if userRole != common.RoleAdmin {
			query += " AND user_id = $2" // Non-admins can only see their own orders
			args = append(args, userID)
		}

		rows, err := app.DB.QueryContext(c, query, args...) // Changed query to filter by user ID
		if err != nil {

			l.DebugF("Error querying orders: %v", err)                                        // Detailed log
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders."}) // Generic error message to client

			return

		}
		defer rows.Close()

		orders := make([]Order, 0) // Make sure orders is initialized.

		for rows.Next() {

			var order Order

			err = rows.Scan(&order.ID, &order.CartID, &order.UserID, &order.AddressID, &order.Total, &order.Updated, &order.Created)

			if err != nil {

				l.DebugF("Error scanning orders : %v", err)

				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders."})

				return

			}

			orders = append(orders, order)

		}

		c.JSON(http.StatusOK, gin.H{"orders": orders})

	}
}

func FetchOrders(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		pageStr := c.DefaultQuery("page", "1")
		limitStr := c.DefaultQuery("limit", "10")

		pageNum, err := strconv.Atoi(pageStr)
		if err != nil || pageNum < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
			return
		}

		limitNum, err := strconv.Atoi(limitStr)
		if err != nil || limitNum < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit"})
			return
		}

		offset := (pageNum - 1) * limitNum

		rows, err := app.DB.QueryContext(c, `
			SELECT id, cart_id, user_id, address_id, total, updated, created
			FROM orders
			ORDER BY created DESC
			LIMIT $1 OFFSET $2
		`, limitNum, offset)
		if err != nil {
			l.ErrorF("Database query error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve orders"})
			return
		}
		defer rows.Close()

		orders := []Order{} // Initialize as empty slice to avoid null in response
		for rows.Next() {
			var order Order
			err := rows.Scan(&order.ID, &order.CartID, &order.UserID, &order.AddressID, &order.Total, &order.Updated, &order.Created)
			if err != nil {
				l.ErrorF("Error scanning order: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan order data"})
				return
			}
			orders = append(orders, order)
		}

		var totalOrders int
		err = app.DB.QueryRowContext(c, "SELECT COUNT(*) FROM orders").Scan(&totalOrders) // Use QueryRowContext for count

		if err != nil {
			l.ErrorF("Error counting total orders: %v", err)                                 // Log the error for debugging
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count orders"}) // Generic message to client
			return
		}

		totalPages := int(math.Ceil(float64(totalOrders) / float64(limitNum)))

		c.JSON(http.StatusOK, gin.H{
			"orders":       orders,
			"total_pages":  totalPages,
			"current_page": pageNum,     // Use "current_page" for consistency
			"total_orders": totalOrders, // Include total order count

		})
	}
}

func FetchUserOrders(app *conf.Config) gin.HandlerFunc { // Update FetchUserOrders
	return func(c *gin.Context) {

		pageStr := c.DefaultQuery("page", "1")
		limitStr := c.DefaultQuery("limit", "10")
		pageNum, err := strconv.Atoi(pageStr)
		if err != nil || pageNum < 1 {

			pageNum = 1 // Set default value if parsing fails or is less than 1

		}
		limitNum, err := strconv.Atoi(limitStr)
		if err != nil || limitNum < 1 {

			limitNum = 10 // Set reasonable default if parsing fails or is less than 1
		}

		userIDStr := c.GetString("userID")
		userID, err := uuid.Parse(userIDStr)
		if err != nil {

			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return

		}

		offset := (pageNum - 1) * limitNum

		rows, err := app.DB.QueryContext(c, `
			SELECT id, cart_id, user_id, address_id, total, updated, created
			FROM orders
			WHERE user_id = $1
			ORDER BY created DESC  -- Order by created timestamp, descending
			LIMIT $2 OFFSET $3
		`, userID, limitNum, offset)

		if err != nil {
			l.ErrorF("Error querying user orders: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
			return
		}
		defer rows.Close() // Ensure rows are closed

		orders := []Order{}
		for rows.Next() {
			var order Order
			if err := rows.Scan(&order.ID, &order.CartID, &order.UserID, &order.AddressID, &order.Total, &order.Updated, &order.Created); err != nil {

				l.ErrorF("Failed to scan user orders: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan order data"})

				return

			}

			orders = append(orders, order)
		}

		// Get total count of orders for this user
		var totalOrders int
		err = app.DB.QueryRowContext(c, "SELECT COUNT(*) FROM orders WHERE user_id = $1", userID).Scan(&totalOrders)
		if err != nil {
			l.ErrorF("Failed to count total user orders: %v", err)                                // Log error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order count"}) // Send generic error message to the client
			return
		}

		totalPages := int(math.Ceil(float64(totalOrders) / float64(limitNum)))

		c.JSON(http.StatusOK, gin.H{
			"orders":       orders,
			"total_pages":  totalPages,
			"page":         pageNum,
			"total_orders": totalOrders, // Return the total order count
		})

	}
}

func FetchOrder(app *conf.Config) gin.HandlerFunc { // Updated FetchOrder
	return func(c *gin.Context) {
		orderIDStr := c.Param("orderId")

		orderID, err := uuid.Parse(orderIDStr)
		if err != nil {

			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID format"})
			return

		}

		userIDStr := c.GetString("userID") // Assuming this is set by middleware
		userID, err := uuid.Parse(userIDStr)
		if err != nil {

			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"}) // Updated error message to be more specific

			return
		}

		userRoleStr := c.MustGet("role").(string)

		userRole := common.GetUserRole(userRoleStr)

		// Use a transaction for consistent reads

		ctx := context.Background()
		tx, err := app.DB.BeginTx(ctx, &sql.TxOptions{ReadOnly: true}) // Use read-only transaction here for fetching
		if err != nil {

			l.ErrorF("Failed to start transaction: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
			return
		}
		defer tx.Rollback()

		// Fetch order. This initializes `order.AddressID` which is needed in the later queries.

		var order Order
		err = tx.QueryRowContext(ctx, `SELECT * FROM orders WHERE id = $1`, orderID).Scan(&order.ID, &order.CartID, &order.UserID, &order.AddressID, &order.Total, &order.Updated, &order.Created)

		// ... (handle error, check for "no rows")

		if err != nil {

			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{"message": "Order not found"})
			} else {

				l.DebugF("Failed to fetch order: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve order"}) // Generic message
			}
			return

		}

		if userRole != common.RoleAdmin && order.UserID != userID {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return

		}

		var orderInfo OrderInfo

		orderInfo.ID = order.ID

		orderInfo.CartID = order.CartID

		orderInfo.UserID = order.UserID

		orderInfo.Total = order.Total

		orderInfo.Updated = order.Updated

		orderInfo.Created = order.Created

		err = tx.QueryRowContext(ctx, `SELECT * from addresses WHERE id=$1`, order.AddressID).Scan(&orderInfo.Address.ID, &orderInfo.Address.UserID, &orderInfo.Address.AddressLine1, &orderInfo.Address.AddressLine2, &orderInfo.Address.City, &orderInfo.Address.State, &orderInfo.Address.Country, &orderInfo.Address.ZipCode, &orderInfo.Address.IsDefault, &orderInfo.Address.Updated, &orderInfo.Address.Created)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {

				c.JSON(http.StatusNotFound, gin.H{"error": "Address associated with this order not found"})
			} else {

				l.ErrorF("Failed to fetch address data: %v", err)                                       // Log the error for debugging
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch order address"}) // Return a generic message for security
			}
			return

		}

		// Fetch associated cart items
		rows, err := tx.QueryContext(ctx, `
			SELECT ci.product_id, ci.quantity, ci.purchase_price, ci.status
			FROM cart_items ci
			WHERE ci.cart_id = $1
		`, order.CartID)
		if err != nil {

			l.DebugF("Error fetching cart items: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cart items"})

			return
		}
		rows.Close()

		// defer rows.Close()                            // Close rows to free resources
		orderInfo.Products = make([]cart.CartItem, 0) // Initialize to empty slice

		for rows.Next() {

			var cartItem cart.CartItem
			err = rows.Scan(&cartItem.ProductID, &cartItem.Quantity, &cartItem.PurchasePrice, &cartItem.Status)

			if err != nil {
				l.DebugF("Failed to scan cart item: %v", err)                                             // Detailed error message
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve cart details"}) // Generic error for security

				return

			}
			l.DebugF("Fetching product details for product ID: %s", cartItem.ProductID)
			var product product.Product
			query := "SELECT id, sku, name, slug, image_url, description, quantity, price, taxable, is_active, brand_id, merchant_id, updated, created FROM products WHERE id = $1"

			err = tx.QueryRowContext(ctx, query, cartItem.ProductID).Scan(
				&product.ID, &product.SKU, &product.Name, &product.Slug, &product.ImageURL, &product.ImageKey,
				&product.Description, &product.Quantity, &product.Price, &product.Taxable, &product.IsActive,
				&product.BrandID, &product.MerchantID, &product.Updated, &product.Created,
			)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Product with ID %s not found in cart", cartItem.ProductID)}) // More informative error
				} else {
					l.ErrorF("Failed to fetch product: %#v", err)                                          // Log with more context
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product data"}) // Generic message for security
				}
				return
			}
			cartItem.Product = &product
			orderInfo.Products = append(orderInfo.Products, cartItem)

		}

		c.JSON(http.StatusOK, gin.H{"order": orderInfo})

	}
}

func CancelOrder(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		orderIDStr := c.Param("orderId")
		orderID, err := uuid.Parse(orderIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
			return
		}

		userIDStr := c.GetString("userID")
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		ctx := context.Background()
		tx, err := app.DB.BeginTx(ctx, nil)
		if err != nil {
			l.ErrorF("Failed to begin transaction: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
			return
		}
		defer tx.Rollback() // Defer rollback

		var cartID uuid.UUID
		err = tx.QueryRowContext(ctx, "SELECT cart_id FROM orders WHERE id = $1 AND user_id = $2", orderID, userID).Scan(&cartID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Order not found or unauthorized"}) // Combined message for security
			} else {
				l.ErrorF("Error fetching cart ID: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve cart ID"})
			}
			return
		}

		// Use ExecContext within the transaction for all DELETE operations:

		_, err = tx.ExecContext(ctx, "DELETE FROM cart_items WHERE cart_id = $1", cartID)
		if err != nil {
			l.ErrorF("Failed to delete cart items: %v", err) // Log the error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete cart items"})
			return
		}

		_, err = tx.ExecContext(ctx, "DELETE FROM carts WHERE id = $1", cartID)
		if err != nil {
			l.ErrorF("Failed to delete cart: %v", err) // Log the error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete cart"})
			return
		}

		_, err = tx.ExecContext(ctx, "DELETE FROM orders WHERE id = $1 AND user_id = $2", orderID, userID)
		if err != nil {
			l.ErrorF("Failed to delete order: %v", err) // Log the error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete order"})
			return
		}

		if err := tx.Commit(); err != nil { // Commit transaction
			l.ErrorF("Transaction commit error: %v", err)                                          // Log error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"}) // Don't reveal DB details
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Order cancelled successfully"})
	}

}

func UpdateItemStatus(app *conf.Config) gin.HandlerFunc { // Updated UpdateItemStatus
	return func(c *gin.Context) {

		orderItemIDStr := c.Param("itemId") // Change parameter name
		orderItemID, err := uuid.Parse(orderItemIDStr)
		if err != nil {

			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order item ID"}) // More specific message

			return
		}

		var req UpdateOrderItemStatusRequest

		if err := c.ShouldBindJSON(&req); err != nil {

			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		userIDStr := c.GetString("userID")
		userID, _ := uuid.Parse(userIDStr)
		userRoleStr := c.MustGet("role").(string)
		userRole := common.GetUserRole(userRoleStr)

		status := req.Status // Use status from request directly
		ctx := context.Background()
		tx, err := app.DB.BeginTx(ctx, nil) // Start transaction
		if err != nil {

			l.DebugF("Error starting transaction: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
			return

		}

		defer tx.Rollback()

		// Check if the order item exists and get details for authorization and updates
		var orderItem OrderItem
		err = tx.QueryRowContext(ctx, `
			SELECT oi.order_id, oi.product_id, oi.quantity, o.user_id  -- Select necessary fields
			FROM order_items oi  -- Correct table name
			JOIN orders o ON oi.order_id = o.id  -- Assuming you have an orders table with user_id
			WHERE oi.id = $1  -- Correct where condition
		`, orderItemID).Scan(&orderItem.OrderID, &orderItem.ProductID, &orderItem.Quantity, &userID) // Get user ID for verification

		if err != nil {

			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{"message": "Order item not found."}) // More user-friendly
			} else {
				l.ErrorF("Failed to get order item details: %v", err)                               // Log for debugging
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get order item."}) // Generic error message
			}

			return

		}

		// Authorization Check: Ensure current user is authorized
		if userRole == common.RoleMerchant {

			var merchantID uuid.UUID

			err = tx.QueryRowContext(ctx, `SELECT merchant_id from products where id=$1`, orderItem.ProductID).Scan(&merchantID)

			if err != nil {
				l.DebugF("Error getting merchant ID: %v", err) // Log with detail

				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve merchant"}) // Return a more generic error

				return

			}

			var authMerchant common.Merchant
			err = tx.QueryRowContext(ctx, "SELECT user_id FROM merchants WHERE id = $1", merchantID).Scan(&authMerchant.UserID)

			if err != nil {
				l.ErrorF("Error fetching merchant details: %v", err)                                 // Log for debug
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify merchant."}) // More generic message

				return

			}

			if authMerchant.UserID != userID {

				c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to update this order item."}) // Correct status code and more specific error

				return

			}

		} else if userRole != common.RoleAdmin { // Admins can change any order status

			c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized access."}) // More specific message

			return

		}

		// Update the cart_items.status within the transaction.  Add error handling.
		_, err = tx.ExecContext(ctx, "UPDATE cart_items SET status = $1, updated = $2 WHERE id = $3", status, time.Now(), orderItemID)

		if err != nil {

			l.ErrorF("Failed to update order item status: %v", err)                                      // Log for debugging.
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order item status"}) // Generic error message to client

			return
		}

		if status == cart.Cancelled { // Use the enum from the correct package

			// Update product quantity
			_, err = tx.ExecContext(ctx, `
                UPDATE products 
                SET quantity = quantity + $1 
                WHERE id = $2
            `, orderItem.Quantity, orderItem.ProductID) // Use tx.ExecContext and correct query
			if err != nil {

				l.DebugF("Failed to update product quantity: %v", err)                                      // Log error
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product quantity"}) // Generic error to client

				return

			}

			// Check if all items are cancelled

			var activeOrderItemsCount int

			err = tx.QueryRowContext(ctx, `
				SELECT COUNT(*) FROM cart_items WHERE cart_id = $1 AND status != $2
			`, req.CartID, cart.Cancelled).Scan(&activeOrderItemsCount) // Use $1 and req.CartID.

			if err != nil {

				l.DebugF("Failed to count cart items: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check cart items."}) // More generic

				return
			}

			if activeOrderItemsCount == 0 {

				_, err = tx.ExecContext(ctx, "DELETE FROM carts WHERE id = $1", req.CartID)
				if err != nil {

					l.DebugF("Failed to delete cart: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete cart"})

					return
				}

				_, err = tx.ExecContext(ctx, "DELETE FROM orders WHERE id = $1", orderItem.OrderID) // Delete the order now

				if err != nil {

					l.DebugF("Failed to delete order: %v", err) // Detailed log

					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel order"}) // Generic error message to the user

					return
				}

				// Return now to avoid the later response
				c.JSON(http.StatusOK, gin.H{"success": true, "orderCancelled": true, "message": "Order has been cancelled!"})

				// Important: commit the transaction after successful cancellation
				if err = tx.Commit(); err != nil {

					l.ErrorF("Error committing transaction: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
					return

				}

				return
			}

		}

		if err := tx.Commit(); err != nil {

			l.ErrorF("Failed to commit transaction: %v", err) // Log the error

			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return

		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Order item status updated successfully"})
	}

}
