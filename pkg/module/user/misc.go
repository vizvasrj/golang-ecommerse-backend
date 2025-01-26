package user

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"src/common"
	"src/l"
	"src/pkg/conf"
)

func CreateMerchantUser(app *conf.Config, ctx context.Context, email, name string, merchantID uuid.UUID) error {
	firstName := name
	lastName := ""

	tx, err := app.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback on error

	var existingUserID uuid.UUID
	err = tx.QueryRowContext(ctx, "SELECT id FROM users WHERE email = $1", email).Scan(&existingUserID)

	if err != nil && !errors.Is(err, sql.ErrNoRows) { // Check if error is NOT "no rows"

		return fmt.Errorf("failed to check for existing user: %w", err)

	}

	if err == nil { // Existing user found
		// Update existing user

		_, err = tx.ExecContext(ctx, `
			UPDATE users 
			SET merchant_id = $1, role = $2, updated = $3 
			WHERE id = $4
		`, merchantID, common.RoleMerchant, time.Now(), existingUserID)
		if err != nil {

			return fmt.Errorf("failed to update existing user: %w", err)
		}

		var merchantExists bool

		err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM merchants WHERE id = $1)`, merchantID).Scan(&merchantExists)
		if err != nil {

			return fmt.Errorf("failed to find merchant: %w", err)
		}

		if !merchantExists {

			return fmt.Errorf("merchant not found with ID: %s", merchantID)

		}

	} else { // New user; create user and send signup email

		newUserID := uuid.New()

		resetToken := generateResetToken()

		_, err = tx.ExecContext(ctx, `
			INSERT INTO users (id, email, first_name, last_name, reset_password_token, merchant_id, role, created)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`, newUserID, email, firstName, lastName, resetToken, merchantID, common.RoleMerchant, time.Now())

		if err != nil {

			return fmt.Errorf("failed to insert new user: %w", err)
		}

	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Email sending should be handled outside the database transaction.

	return nil

}

func generateResetToken() string { // Helper function (same as before)
	buffer := make([]byte, 48)
	_, err := rand.Read(buffer)
	if err != nil {
		l.ErrorF("Failed to generate random token: %v", err)
		return "" // Or handle the error more gracefully
	}
	return hex.EncodeToString(buffer)
}
