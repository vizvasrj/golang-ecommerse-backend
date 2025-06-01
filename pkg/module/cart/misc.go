package cart

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

// checkCartOwnership validates cart access rights
func checkCartOwnership(tx *sql.Tx, ctx context.Context, cartID uuid.UUID, userID uuid.UUID) (bool, error) {
	var dbUserID uuid.NullUUID
	err := tx.QueryRowContext(ctx,
		"SELECT user_id FROM carts WHERE id = $1",
		cartID,
	).Scan(&dbUserID)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil // Cart doesn't exist
		}
		return false, err
	}

	// Authenticated user must match cart's user_id
	if userID != uuid.Nil {
		return dbUserID.Valid && dbUserID.UUID == userID, nil
	}

	// Anonymous user must have NULL user_id in cart
	return !dbUserID.Valid, nil
}

// updateCartItem handles item insertion/update
func updateCartItem(tx *sql.Tx, ctx context.Context, cartID, productID uuid.UUID, quantity int, action string) error {
	// Verify product exists and get price
	var price float64
	err := tx.QueryRowContext(ctx,
		"SELECT price FROM products WHERE id = $1",
		productID,
	).Scan(&price)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("product not found")
		}
		return err
	}

	// Check existing quantity
	var existingQty int
	err = tx.QueryRowContext(ctx, `
		SELECT quantity FROM cart_items 
		WHERE cart_id = $1 AND product_id = $2
	`, cartID, productID).Scan(&existingQty)

	switch {
	case err == sql.ErrNoRows:
		// Insert new item
		_, err = tx.ExecContext(ctx, `
			INSERT INTO cart_items (cart_id, product_id, quantity, purchase_price)
			VALUES ($1, $2, $3, $4)
		`, cartID, productID, quantity, price)
		return err

	case err != nil:
		return err

	default:
		// Update existing item
		newQty := existingQty + quantity
		if action == "replace" {
			newQty = quantity
		}

		_, err = tx.ExecContext(ctx, `
			UPDATE cart_items 
			SET quantity = $1, updated = NOW()
			WHERE cart_id = $2 AND product_id = $3
		`, newQty, cartID, productID)
		return err
	}
}
