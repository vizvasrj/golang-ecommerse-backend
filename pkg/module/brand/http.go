package brand

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"src/l"
	"src/pkg/conf"
)

// HTTP Handlers (Updated for Postgres and sqlx)

func AddBrand(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var brand Brand // Use the Brand struct

		if err := c.ShouldBindJSON(&brand); err != nil {
			l.DebugF("Binding error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		if brand.Name == "" || brand.Description == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Name and description are required"})
			return
		}

		newBrandID := uuid.New()

		_, err := app.DB.ExecContext(c, `
			INSERT INTO brands (id, name, slug, image, content_type, description, is_active, updated, created)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, newBrandID, brand.Name, brand.Slug, brand.Image, brand.ContentType, brand.Description, brand.IsActive, time.Now(), time.Now())
		if err != nil {

			l.DebugF("Error inserting brand: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add brand"})
			return
		}

		brand.ID = newBrandID // Set the ID

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Brand has been added successfully!", "brand": brand})

	}
}

func ListBrands(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := `
		SELECT 
			id, name, slug, image, content_type, description, is_active, updated, created
		FROM brands WHERE is_active = TRUE
		`
		rows, err := app.DB.QueryContext(c, query)
		if err != nil {
			l.ErrorF("Error querying brands: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch brands"})
			return
		}
		defer rows.Close()

		var brands []Brand
		for rows.Next() {
			var fetchedBrand Brand
			if err := rows.Scan(&fetchedBrand.ID, &fetchedBrand.Name, &fetchedBrand.Slug, &fetchedBrand.Image, &fetchedBrand.ContentType, &fetchedBrand.Description, &fetchedBrand.IsActive, &fetchedBrand.Updated, &fetchedBrand.Created); err != nil {
				l.ErrorF("Error scanning brand: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch brands"})
				return
			}
			brands = append(brands, fetchedBrand)
		}

		c.JSON(http.StatusOK, gin.H{"brands": brands})
	}
}

//GetBrands functions does same like ListBrands() so just re-use ListBrands

// get brand by ID
func GetBrandByID(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {

		brandIDStr := c.Param("id")
		brandID, err := uuid.Parse(brandIDStr)
		if err != nil {
			l.DebugF("Invalid brand ID: %v", err) // Log the error
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid brand ID"})
			return
		}

		var fetchedBrand Brand
		query := `
			SELECT
				id, name, slug, image, content_type, description, is_active, updated, created
			FROM brands
			WHERE id = $1


		`
		err = app.DB.QueryRowContext(c, query, brandID).Scan(&fetchedBrand.ID, &fetchedBrand.Name, &fetchedBrand.Slug, &fetchedBrand.Image, &fetchedBrand.ContentType, &fetchedBrand.Description, &fetchedBrand.IsActive, &fetchedBrand.Updated, &fetchedBrand.Created)
		if err != nil {

			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{"message": "Brand not found"})
			} else {
				l.DebugF("Database query error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch brand"})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"brand": fetchedBrand})
	}
}

func ListSelectBrands(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := app.DB.QueryContext(c, "SELECT id, name FROM brands") // Select only necessary fields
		if err != nil {
			l.ErrorF("Error querying brands: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch brands"})
			return
		}
		defer rows.Close()

		var brands []Brand
		for rows.Next() {
			var fetchedBrand Brand
			if err := rows.Scan(&fetchedBrand.ID, &fetchedBrand.Name); err != nil {
				l.ErrorF("Error scanning brand: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch brands"})
				return
			}
			brands = append(brands, fetchedBrand)
		}

		c.JSON(http.StatusOK, gin.H{"brands": brands})
	}
}

func UpdateBrand(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		brandIDStr := c.Param("id")
		brandID, err := uuid.Parse(brandIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid brand ID"})
			return
		}

		var updateBrand BrandUpdate
		if err := c.ShouldBindJSON(&updateBrand); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var updateFields []string
		var args []interface{}

		// Check slug uniqueness if being updated
		if updateBrand.Slug != nil {
			var existingBrandID uuid.UUID
			err := app.DB.QueryRowContext(c, `SELECT id FROM brands WHERE slug = $1`, *updateBrand.Slug).Scan(&existingBrandID)
			if err == nil && existingBrandID != brandID {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Slug is already in use."})
				return
			} else if err != nil && err != sql.ErrNoRows {
				l.DebugF("Error checking slug uniqueness: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update brand."})
				return
			}
		}

		// Append fields dynamically
		args = append(args, time.Now())
		updateFields = append(updateFields, "updated = $1")

		if updateBrand.Name != nil {
			updateFields = append(updateFields, fmt.Sprintf("name = $%d", len(args)+1))
			args = append(args, *updateBrand.Name)
		}
		if updateBrand.Slug != nil {
			updateFields = append(updateFields, fmt.Sprintf("slug = $%d", len(args)+1))
			args = append(args, *updateBrand.Slug)
		}
		if updateBrand.Image != nil {
			updateFields = append(updateFields, fmt.Sprintf("image = $%d", len(args)+1))
			args = append(args, *updateBrand.Image)
		}
		if updateBrand.ContentType != nil {
			updateFields = append(updateFields, fmt.Sprintf("content_type = $%d", len(args)+1))
			args = append(args, *updateBrand.ContentType)
		}
		if updateBrand.Description != nil {
			updateFields = append(updateFields, fmt.Sprintf("description = $%d", len(args)+1))
			args = append(args, *updateBrand.Description)
		}
		if updateBrand.IsActive != nil {
			updateFields = append(updateFields, fmt.Sprintf("is_active = $%d", len(args)+1))
			args = append(args, *updateBrand.IsActive)
		}

		// If no fields to update, return early
		if len(updateFields) == 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
			return
		}

		// Add brandID for WHERE clause
		updateQuery := fmt.Sprintf("UPDATE brands SET %s WHERE id = $%d",
			strings.Join(updateFields, ", "), len(args)+1)
		args = append(args, brandID)

		// Check for slug uniqueness if it's being updated
		if updateBrand.Slug != nil {
			var existingBrandID uuid.UUID
			err := app.DB.QueryRowContext(c, `SELECT id FROM brands WHERE slug = $1`, *updateBrand.Slug).Scan(&existingBrandID)
			if err == nil && existingBrandID != brandID {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Slug is already in use."})
				return
			} else if err != nil && err != sql.ErrNoRows {
				l.DebugF("Error checking slug uniqueness: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update brand."})
				return
			}
		}

		result, err := app.DB.ExecContext(c, updateQuery, args...)
		if err != nil {

			l.DebugF("Error updating brand: %v", err) // Log the error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update brand"})
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			l.DebugF("Error getting affected rows: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update brand"})
			return
		}

		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Brand not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Brand updated successfully"})
	}
}

func UpdateBrandActive(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		brandIDStr := c.Param("id")
		brandID, err := uuid.Parse(brandIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid brand ID"})
			return
		}

		var updateBrand struct {
			IsActive *bool `json:"isActive"`
		}
		if err := c.ShouldBindJSON(&updateBrand); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		if updateBrand.IsActive == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "isActive field is required"})
			return
		}
		_, err = app.DB.ExecContext(c, `
			UPDATE brands 
			SET is_active = $1, updated = $2  
			WHERE id = $3
		`, updateBrand.IsActive, time.Now(), brandID)
		if err != nil {
			l.DebugF("Error updating brand status: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update brand status"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Brand status updated successfully"})
	}
}

func DeleteBrand(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		brandIDStr := c.Param("id")
		brandID, err := uuid.Parse(brandIDStr)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid brand ID"})
			return
		}

		result, err := app.DB.ExecContext(c, "DELETE FROM brands WHERE id = $1", brandID)
		if err != nil {
			l.DebugF("Error deleting brand: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete brand"})
			return
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			l.DebugF("Error getting rows affected: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete brand"})
			return
		}

		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"message": "Brand not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Brand deleted successfully"})

	}
}
