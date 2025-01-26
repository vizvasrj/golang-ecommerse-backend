package cart

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"src/l"
	"src/pkg/conf"
	"src/pkg/module/product"
)

func AddToCart(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var cartProduct AddProductToCartRequest
		if err := c.ShouldBindJSON(&cartProduct); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
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
		defer tx.Rollback() // Defer rollback in case of any errors

		newCartID := uuid.New()
		_, err = tx.ExecContext(ctx, "INSERT INTO carts (id, user_id, created, updated) VALUES ($1, $2, $3, $4)", newCartID, userID, time.Now(), time.Now())
		if err != nil {
			l.ErrorF("Error creating cart: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create cart"})
			return
		}

		newCartItemID := uuid.New()

		_, err = tx.ExecContext(ctx, `
			INSERT INTO cart_items (id, cart_id, product_id, quantity, purchase_price, created, updated)
			SELECT $1, $2, $3, $4, p.price, $5, $6  -- Get price directly from products table
			FROM products p
			WHERE p.id = $3
		`, newCartItemID, newCartID, cartProduct.ProductID, cartProduct.Quantity, time.Now(), time.Now())
		if err != nil {
			l.ErrorF("Error adding item to cart: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add item to cart"})
			return
		}

		if err := tx.Commit(); err != nil { // Commit transaction

			l.ErrorF("Error committing transaction: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "cart_id": newCartID})

	}
}

// DeleteCart function (updated below)
// AddProductToCart function (updated below)

func DeleteCart(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		cartIDStr := c.Param("cartId")
		cartID, err := uuid.Parse(cartIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cart ID"})
			return
		}

		userIDStr := c.GetString("userID")
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		ctx := context.Background()
		tx, err := app.DB.BeginTx(ctx, nil) // Start transaction
		if err != nil {
			l.ErrorF("Error beginning transaction: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
			return
		}
		defer tx.Rollback()

		var cartExists bool
		err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM carts WHERE id = $1 AND user_id = $2)", cartID, userID).Scan(&cartExists)

		if err != nil {
			l.ErrorF("Error checking cart existence: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check cart ownership"})
			return
		}
		if !cartExists {

			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
			return
		}

		_, err = tx.ExecContext(ctx, "DELETE FROM cart_items WHERE cart_id = $1", cartID) // Corrected table name
		if err != nil {
			l.ErrorF("Error deleting cart items: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete cart items"})
			return
		}

		_, err = tx.ExecContext(ctx, "DELETE FROM carts WHERE id = $1 AND user_id = $2", cartID, userID)
		if err != nil {
			l.ErrorF("Error deleting cart: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete cart"})
			return
		}

		if err := tx.Commit(); err != nil { // Commit transaction
			l.ErrorF("Error committing transaction: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true})

	}
}

func AddProductToCart(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		cartIDStr := c.Param("cartId")
		cartID, err := uuid.Parse(cartIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cart ID"})
			return
		}

		userIDStr := c.GetString("userID")
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		var cartItem CartItemRequest // Updated struct name
		if err := c.ShouldBindJSON(&cartItem); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cart item data"})
			return
		}

		ctx := context.Background()
		tx, err := app.DB.BeginTx(ctx, nil) // Start transaction
		if err != nil {
			l.ErrorF("Error beginning transaction: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
			return
		}
		defer tx.Rollback() // Defer rollback

		// Check if cart exists and belongs to the user
		var cartExists bool

		err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM carts WHERE id = $1 AND user_id = $2)", cartID, userID).Scan(&cartExists)

		if err != nil {

			l.ErrorF("Error checking cart existence: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify cart ownership"})
			return
		}

		if !cartExists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
			return
		}

		// Check if the product already exists in the cart
		var existingCartItem CartItem
		err = tx.QueryRowContext(ctx, `SELECT id, quantity FROM cart_items WHERE cart_id = $1 AND product_id = $2`, cartID, cartItem.ProductID).Scan(&existingCartItem.ID, &existingCartItem.Quantity)

		if err != nil && err != sql.ErrNoRows {
			l.DebugF("Error checking for existing cart item: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check for existing item in cart"})
			return
		}

		if err == sql.ErrNoRows {
			// Product doesn't exist in cart, insert new item

			newCartItemID := uuid.New()
			_, err = tx.ExecContext(ctx, `
				INSERT INTO cart_items (id, cart_id, product_id, quantity, purchase_price, created, updated)
				SELECT $1, $2, $3, $4, p.price, $5, $6 
				FROM products p
				WHERE p.id = $3
			`, newCartItemID, cartID, cartItem.ProductID, cartItem.Quantity, time.Now(), time.Now())
			if err != nil {
				l.ErrorF("Error inserting new cart item: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add item to cart"})
				return
			}
		} else {
			// Product exists in cart, update quantity

			_, err = tx.ExecContext(ctx, "UPDATE cart_items SET quantity = quantity + $1, updated = $2 WHERE id = $3", cartItem.Quantity, time.Now(), existingCartItem.ID)
			if err != nil {
				l.ErrorF("Error updating cart item quantity: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cart"})
				return
			}
		}

		if err := tx.Commit(); err != nil { // Commit transaction
			l.ErrorF("Error committing transaction: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}

func RemoveProductFromCart(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		cartIDStr := c.Param("cartId")
		cartID, err := uuid.Parse(cartIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cart ID"})
			return
		}

		userIDStr := c.GetString("userID")
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		productIDStr := c.Param("productId")
		productID, err := uuid.Parse(productIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
			return
		}

		ctx := context.Background()
		tx, err := app.DB.BeginTx(ctx, nil) // Start a transaction
		if err != nil {
			l.ErrorF("Transaction start failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
			return
		}
		defer tx.Rollback() // Ensure rollback on error

		// Verify cart ownership within the transaction
		var cartExists bool
		err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM carts WHERE id = $1 AND user_id = $2)", cartID, userID).Scan(&cartExists)
		if err != nil {
			l.ErrorF("Error checking for cart: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify cart ownership"})
			return
		}
		if !cartExists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		result, err := tx.ExecContext(ctx, "DELETE FROM cart_items WHERE cart_id = $1 AND product_id = $2", cartID, productID)

		if err != nil {
			l.ErrorF("Error removing product from cart: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove product from cart"})
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			l.ErrorF("Error getting rows affected: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get rows affected"})
			return
		}
		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found in cart"})
			return
		}

		if err := tx.Commit(); err != nil {
			l.ErrorF("Transaction commit failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Product removed from cart"})
	}
}

func AddProductToCartV2(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract cart ID, create a new one if it doesn't exist
		cartID, err := uuid.Parse(c.Query("cartId"))
		if err != nil {
			l.DebugF("Error parsing cart ID: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cart ID"})
			return
		}
		// ... (userID and cartItem parsing - same as before)

		userIDStr := c.GetString("userID")
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		var cartItem CartItemRequest // Correct name
		if err := c.ShouldBindJSON(&cartItem); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product data"})
			return
		}

		// ... (verify product exists - update function to use sqlx and UUID)

		var productExists bool

		err = app.DB.QueryRow("SELECT EXISTS (SELECT 1 FROM products WHERE id = $1)", cartItem.ProductID).Scan(&productExists)
		if err != nil {

			l.ErrorF("Error checking product existence: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify product existence"})
			return
		}

		if !productExists {

			c.JSON(http.StatusBadRequest, gin.H{"error": "Product not found"})
			return
		}

		ctx := context.Background()
		tx, err := app.DB.BeginTx(ctx, nil)
		if err != nil {

			l.ErrorF("Error beginning transaction: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
			return
		}
		defer tx.Rollback() // Defer transaction rollback

		if cartID == uuid.Nil {
			// Create new cart

			cartID = uuid.New()
			_, err = tx.ExecContext(ctx, "INSERT INTO carts (id, user_id, created, updated) VALUES ($1, $2, $3, $4)", cartID, userID, time.Now(), time.Now())

			if err != nil {

				l.ErrorF("Error creating cart: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create cart"})
				return
			}

			// Insert cart item
			_, err = tx.ExecContext(ctx, `INSERT INTO cart_items (cart_id, product_id, quantity, purchase_price, created, updated)
											SELECT $1, $2, $3, p.price, $4, $5
											FROM products p
											WHERE p.id = $2`, cartID, cartItem.ProductID, cartItem.Quantity, time.Now(), time.Now())

			if err != nil {
				l.ErrorF("Error adding item to new cart: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add item to cart"})
				return
			}

		} else {

			var cart Cart
			err = tx.QueryRowContext(ctx, "SELECT id, user_id FROM carts WHERE id = $1", cartID).Scan(&cart.ID, &cart.UserID)

			if err != nil {
				if err == sql.ErrNoRows {
					c.JSON(http.StatusNotFound, gin.H{"error": "Cart not found"})

				} else {

					l.ErrorF("Error fetching cart: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cart"})
				}
				return
			}

			if cart.UserID != userID {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
				return
			}

			// Check if item exists to update or insert

			var existingCartItem CartItem

			err = tx.QueryRowContext(ctx, `
				SELECT id, quantity 
				FROM cart_items 
				WHERE cart_id = $1 AND product_id = $2
			`, cartID, cartItem.ProductID).Scan(&existingCartItem.ID, &existingCartItem.Quantity)

			if err != nil {

				if err == sql.ErrNoRows {

					_, err = tx.ExecContext(ctx, `
						INSERT INTO cart_items (cart_id, product_id, quantity, purchase_price, created, updated)
						SELECT $1, $2, $3, p.price, $4, $5
						FROM products p
						WHERE p.id = $2
					`, cartID, cartItem.ProductID, cartItem.Quantity, time.Now(), time.Now())
					if err != nil {
						// Handle insert error
						l.ErrorF("Error inserting new cart item: %v", err)
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add new item to cart"})
						return
					}

				} else {

					l.ErrorF("Error fetching cart item: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch existing cart item"})
					return
				}

			} else {
				// Update existing item
				// TODO if there is increment or replace do as it so/else
				if cartItem.Action == "replace" {
					_, err := tx.ExecContext(ctx, `
						UPDATE cart_items
						SET quantity = $1, updated = $2
						WHERE id = $3
					`, cartItem.Quantity, time.Now(), existingCartItem.ID)
					if err != nil {
						l.ErrorF("Error updating cart item: %v", err)
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cart item"})
						return
					}
				} else {
					_, err := tx.ExecContext(ctx, `
						UPDATE cart_items
						SET quantity = quantity + $1, updated = $2
						WHERE id = $3
					`, cartItem.Quantity, time.Now(), existingCartItem.ID)
					if err != nil {
						l.ErrorF("Error updating cart item: %v", err)
						c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cart item"})
						return
					}

				}
			}
		}

		if err := tx.Commit(); err != nil {
			l.ErrorF("Error committing transaction: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "cart_id": cartID.String()})
	}
}

// ... GetCartByCartID and other functions (updated below)

func GetCartByCartID(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		cartIDStr := c.Param("cartId")
		cartID, err := uuid.Parse(cartIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cart ID"})
			return
		}

		userIDStr := c.GetString("userID") // Get the user ID if available
		var userID uuid.UUID
		if userIDStr != "" {
			userID, err = uuid.Parse(userIDStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
				return
			}
		}

		// Use a transaction to ensure consistent reads
		ctx := context.Background()
		tx, err := app.DB.BeginTx(ctx, &sql.TxOptions{ReadOnly: true}) // Read-only transaction
		if err != nil {
			l.ErrorF("Transaction begin error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
			return
		}
		defer tx.Rollback() // Rollback is a no-op for read-only transactions but good practice

		var cartExists bool
		err = tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM carts WHERE id = $1)", cartID).Scan(&cartExists)

		if err != nil {
			l.ErrorF("Error checking if cart exists: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check cart"})
			return
		}

		if !cartExists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Cart not found"})
			return
		}

		// Check if a user is logged in and if the cart belongs to them
		if userID != uuid.Nil {
			var cartUserID uuid.UUID
			err = tx.QueryRowContext(ctx, "SELECT user_id FROM carts WHERE id = $1", cartID).Scan(&cartUserID)
			if err != nil {
				l.ErrorF("Error fetching cart user ID: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify cart"})
				return
			}
			if cartUserID != userID {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
				return
			}
		}

		rows, err := tx.QueryContext(ctx, `
			SELECT 
				p.id, p.sku, p.name, p.slug, p.image_url, p.description, p.quantity AS product_quantity, p.price,
				ci.quantity AS cart_item_quantity
			FROM products p
			JOIN cart_items ci ON p.id = ci.product_id
			WHERE ci.cart_id = $1
		`, cartID)
		if err != nil {
			l.ErrorF("Error fetching cart items and products: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve cart items"})
			return
		}
		defer rows.Close()

		type CartProduct struct {
			product.Product      // Embed the product struct
			CartItemQuantity int `db:"cart_item_quantity" json:"cart_item_quantity"` // Add the quantity from cart_items
		}

		cartProducts := []CartProduct{}
		for rows.Next() {
			var cartProduct CartProduct
			err = rows.Scan(&cartProduct.ID, &cartProduct.SKU, &cartProduct.Name, &cartProduct.Slug, &cartProduct.ImageURL, &cartProduct.Description, &cartProduct.Quantity, &cartProduct.Price, &cartProduct.CartItemQuantity)
			if err != nil {
				l.ErrorF("Error scanning cart items: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve cart items"})
				return
			}
			cartProducts = append(cartProducts, cartProduct)
		}

		c.JSON(http.StatusOK, gin.H{"cart": cartProducts})

	}
}
