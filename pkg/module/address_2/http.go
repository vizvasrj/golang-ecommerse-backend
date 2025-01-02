package address2

import (
	"database/sql"
	"fmt"
	"net/http"
	"src/l"
	"src/pkg/conf"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid" // Import uuid package
)

// Address represents the address model for Postgres
type Address struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	AddressLine1 string    `json:"address_line1"`
	AddressLine2 string    `json:"address_line2"`
	Address      string    `json:"address"`
	City         string    `json:"city"`
	State        string    `json:"state"`
	Country      string    `json:"country"`
	ZipCode      string    `json:"zip_code"`
	IsDefault    bool      `json:"is_default"`
	Updated      time.Time `json:"updated"`
	Created      time.Time `json:"created"`
}

// AddressAdd represents the request body for adding an address
type AddressAdd struct {
	AddressLine1 string `json:"address_line1"`
	AddressLine2 string `json:"address_line2"`
	City         string `json:"city" binding:"required"`
	State        string `json:"state" binding:"required"`
	Country      string `json:"country" binding:"required"`
	ZipCode      string `json:"zip_code" binding:"required"`
	IsDefault    bool   `json:"is_default"`
}

// AddressUpdate  represents the request body for updating an address
type AddressUpdate struct {
	AddressLine1 string `json:"address_line1"`
	AddressLine2 string `json:"address_line2"`
	City         string `json:"city"`
	State        string `json:"state"`
	Country      string `json:"country"`
	ZipCode      string `json:"zip_code"`
	IsDefault    *bool  `json:"is_default"`
}

func AddAddress(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var address AddressAdd
		if err := c.ShouldBindJSON(&address); err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		userIDStr := c.MustGet("userID").(string) // userID is now a string
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			l.DebugF("Error parsing UUID: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		newAddress := Address{
			ID:           uuid.New(),
			UserID:       userID,
			AddressLine1: address.AddressLine1,
			AddressLine2: address.AddressLine2,
			City:         address.City,
			State:        address.State,
			Country:      address.Country,
			ZipCode:      address.ZipCode,
			IsDefault:    address.IsDefault,
			Updated:      time.Now(),
			Created:      time.Now(),
		}

		_, err = app.DB.Exec("INSERT INTO addresses (id, user_id, address_line1, address_line2, city, state, country, zip_code, is_default, updated, created) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)",
			newAddress.ID, newAddress.UserID, newAddress.AddressLine1, newAddress.AddressLine2, newAddress.City, newAddress.State, newAddress.Country, newAddress.ZipCode, newAddress.IsDefault, newAddress.Updated, newAddress.Created)

		if err != nil {
			l.DebugF("Error inserting address: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not add address"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Address has been added successfully!", "address": address})
	}
}

func GetAddresses(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.MustGet("userID").(string)
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			l.DebugF("Error parsing UUID: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		rows, err := app.DB.Query("SELECT id, user_id, address_line1, address_line2, city, state, country, zip_code, is_default, updated, created FROM addresses WHERE user_id = $1", userID)

		if err != nil {
			l.DebugF("Error querying addresses: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch addresses"})
			return
		}
		defer rows.Close()

		var addresses []Address
		for rows.Next() {
			var address Address
			err := rows.Scan(&address.ID, &address.UserID, &address.AddressLine1, &address.AddressLine2, &address.City, &address.State, &address.Country, &address.ZipCode, &address.IsDefault, &address.Updated, &address.Created)
			if err != nil {
				l.DebugF("Error scanning address: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch addresses"})
				return
			}
			addresses = append(addresses, address)
		}

		c.JSON(http.StatusOK, gin.H{"addresses": addresses})
	}
}

func GetAddress(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		addressIDStr := c.Param("id")
		addressID, err := uuid.Parse(addressIDStr)
		if err != nil {
			l.DebugF("Error parsing address UUID: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address ID"})
			return
		}

		userIDStr := c.MustGet("userID").(string)
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			l.DebugF("Error parsing user UUID: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		var address Address
		err = app.DB.QueryRow("SELECT id, user_id, address_line1,address_line2,  city, state, country, zip_code, is_default, updated, created FROM addresses WHERE id = $1 AND user_id = $2", addressID, userID).
			Scan(&address.ID, &address.UserID, &address.AddressLine1, &address.AddressLine2, &address.City, &address.State, &address.Country, &address.ZipCode, &address.IsDefault, &address.Updated, &address.Created)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"message": "Address not found"})
			return
		} else if err != nil {
			l.DebugF("Error querying address: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch address"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"address": address})
	}
}

func UpdateAddress(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		addressIDStr := c.Param("id")
		addressID, err := uuid.Parse(addressIDStr)
		if err != nil {
			l.DebugF("Error parsing address UUID: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address ID"})
			return
		}

		userIDStr := c.MustGet("userID").(string)
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			l.DebugF("Error parsing user UUID: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		var addressUpdate AddressUpdate // Use the update-specific struct
		if err := c.ShouldBindJSON(&addressUpdate); err != nil {
			l.DebugF("Error binding JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Dynamically build the update query based on provided fields
		updateQuery := "UPDATE addresses SET updated = $1"
		updateArgs := []interface{}{time.Now()}
		counter := 2

		if addressUpdate.AddressLine1 != "" {
			updateQuery += fmt.Sprintf(", address_line1 = $%d", counter)
			updateArgs = append(updateArgs, addressUpdate.AddressLine1)
			counter++
		}
		if addressUpdate.AddressLine2 != "" {
			updateQuery += fmt.Sprintf(", address_line2 = $%d", counter)
			updateArgs = append(updateArgs, addressUpdate.AddressLine2)
			counter++
		}
		if addressUpdate.City != "" {
			updateQuery += fmt.Sprintf(", city = $%d", counter)
			updateArgs = append(updateArgs, addressUpdate.City)
			counter++
		}
		if addressUpdate.State != "" {
			updateQuery += fmt.Sprintf(", state = $%d", counter)
			updateArgs = append(updateArgs, addressUpdate.State)
			counter++
		}
		if addressUpdate.Country != "" {
			updateQuery += fmt.Sprintf(", country = $%d", counter)
			updateArgs = append(updateArgs, addressUpdate.Country)
			counter++
		}
		if addressUpdate.ZipCode != "" {
			updateQuery += fmt.Sprintf(", zip_code = $%d", counter)
			updateArgs = append(updateArgs, addressUpdate.ZipCode)
			counter++
		}
		if addressUpdate.IsDefault != nil {
			updateQuery += fmt.Sprintf(", is_default = $%d", counter)
			updateArgs = append(updateArgs, addressUpdate.IsDefault)
			counter++
		}

		updateQuery += fmt.Sprintf(" WHERE id = $%d AND user_id = $%d", counter, counter+1)
		updateArgs = append(updateArgs, addressID, userID)

		updateQuery += " RETURNING id, user_id, address_line1, address_line2, city, state, country, zip_code, is_default, updated, created"

		var address Address
		err = app.DB.QueryRow(updateQuery, updateArgs...).Scan(&address.ID, &address.UserID, &address.AddressLine1, &address.AddressLine2, &address.City, &address.State, &address.Country, &address.ZipCode, &address.IsDefault, &address.Updated, &address.Created)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Address not found or does not belong to the user"})
			return
		} else if err != nil {
			l.DebugF("Error updating address: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update address"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Address has been updated successfully!", "address": address})
	}
}

// DeleteAddress handles the deletion of an address.
func DeleteAddress(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		addressIDStr := c.Param("id")
		addressID, err := uuid.Parse(addressIDStr)
		if err != nil {
			l.DebugF("Invalid address ID: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address ID"})
			return
		}

		userIDStr := c.MustGet("userID").(string)
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			l.DebugF("Invalid user ID: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		result, err := app.DB.Exec("DELETE FROM addresses WHERE id = $1 AND user_id = $2", addressID, userID)
		if err != nil {
			l.DebugF("Error deleting address: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete address"})
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			l.DebugF("Error getting rows affected: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete address"})
			return
		}

		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Address not found or does not belong to user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Address has been deleted successfully!"})
	}
}

func SetDefaultAddress(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		addressIDStr := c.Param("id")
		addressID, err := uuid.Parse(addressIDStr)
		if err != nil {
			l.DebugF("Error parsing address ID: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address ID"})
			return
		}

		userIDStr := c.MustGet("userID").(string)
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			l.DebugF("Error parsing user ID: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		tx, err := app.DB.Begin()
		if err != nil {
			l.DebugF("Error starting transaction: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not start transaction"})
			return
		}

		defer func() {
			if p := recover(); p != nil {
				tx.Rollback()
				l.DebugF("Recovered from panic: %v", p)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "An unexpected error occurred"})
			}
		}()

		result, err := tx.Exec("UPDATE addresses SET is_default = FALSE WHERE user_id = $1", userID)
		if err != nil {
			l.DebugF("Error updating default address: %v", err)
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update addresses"})
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			l.DebugF("Error fetching rows affected: %v", err)
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update addresses"})
			return
		}

		if rowsAffected == 0 {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "No addresses found for the user"})
			return
		}

		result, err = tx.Exec("UPDATE addresses SET is_default = TRUE, updated = $1 WHERE id = $2 AND user_id = $3", time.Now(), addressID, userID)
		if err != nil {
			l.DebugF("Error setting new default address: %v", err)
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update address"})
			return
		}

		rowsAffected, err = result.RowsAffected()
		if err != nil {
			l.DebugF("Error fetching rows affected: %v", err)
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update address"})
			return
		}

		if rowsAffected == 0 {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "Address not found or does not belong to the user"})
			return
		}

		err = tx.Commit()
		if err != nil {
			l.DebugF("Error committing transaction: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not commit transaction"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Address has been set as default successfully!"})
	}
}
