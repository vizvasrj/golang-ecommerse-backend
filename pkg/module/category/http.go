package category

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"

	"src/l"
	"src/pkg/conf"
)

func AddCategory(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req Category // Use the Category struct directly
		if err := c.ShouldBindJSON(&req); err != nil {

			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		if req.Name == "" || req.Description == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Name and description are required"})
			return
		}

		newCategoryID := uuid.New()

		_, err := app.DB.ExecContext(c, `
			INSERT INTO categories (id, name, slug, description, is_active, updated, created)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, newCategoryID, req.Name, req.Slug, req.Description, req.IsActive, time.Now(), time.Now())

		if err != nil {
			l.DebugF("Error inserting category: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add category"})
			return
		}

		req.ID = newCategoryID

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Category added successfully", "category": req})
	}
}

func ListCategories(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {

		rows, err := app.DB.QueryContext(c, `SELECT 
			id, name, slug, description, is_active, updated, created  
			FROM categories WHERE is_active = TRUE`)
		if err != nil {
			l.DebugF("Error querying categories: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
			return
		}
		defer rows.Close()

		categories := []Category{}
		for rows.Next() {
			var category Category
			if err := rows.Scan(&category.ID, &category.Name, &category.Slug, &category.Description, &category.IsActive, &category.Updated, &category.Created); err != nil {
				l.DebugF("Error scanning category: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
				return
			}

			categories = append(categories, category)
		}

		c.JSON(http.StatusOK, gin.H{"categories": categories})
	}
}

func FetchCategories(app *conf.Config) gin.HandlerFunc { // ... similar to ListCategories, remove is_active = TRUE filter }
	return func(c *gin.Context) {

		rows, err := app.DB.QueryContext(c, "SELECT id, name, slug, description, is_active, updated, created FROM categories")
		if err != nil {
			l.DebugF("Error querying categories: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
			return
		}
		defer rows.Close()

		categories := []Category{}
		for rows.Next() {
			var category Category
			if err := rows.Scan(&category.ID, &category.Name, &category.Slug, &category.Description, &category.IsActive, &category.Updated, &category.Created); err != nil {
				l.DebugF("Error scanning category: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
				return
			}

			categories = append(categories, category)
		}

		c.JSON(http.StatusOK, gin.H{"categories": categories})
	}
}

func FetchCategory(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		categoryIDStr := c.Param("id")
		categoryID, err := uuid.Parse(categoryIDStr)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
			return
		}

		var category Category

		err = app.DB.QueryRowContext(c, "SELECT id, name, slug, description, is_active, updated, created FROM categories WHERE id = $1", categoryID).
			Scan(&category.ID, &category.Name, &category.Slug, &category.Description, &category.IsActive, &category.Updated, &category.Created)
		if err != nil {

			if errors.Is(err, sql.ErrNoRows) { // Correct error check

				c.JSON(http.StatusNotFound, gin.H{"message": "Category not found"})
			} else {
				l.ErrorF("Failed to fetch category: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch category"})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"category": category})

	}
}

func UpdateCategory(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		categoryIDStr := c.Param("id")
		categoryID, err := uuid.Parse(categoryIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
			return
		}

		var updateData CategoryUpdate
		if err := c.ShouldBindJSON(&updateData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		// Start a transaction
		ctx := context.Background()
		tx, err := app.DB.BeginTx(ctx, nil)
		if err != nil {
			l.ErrorF("Failed to begin transaction: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
			return
		}
		defer tx.Rollback() // Defer rollback in case of errors

		// Create dynamic update query and arguments
		updateQuery := "UPDATE categories SET updated = $1"
		args := []interface{}{time.Now()}

		if updateData.Name != nil {
			updateQuery += ", name = $2"
			args = append(args, *updateData.Name)
		}

		if updateData.Slug != nil {
			updateQuery += ", slug = $3"
			args = append(args, *updateData.Slug)

			var existingSlugCount int
			err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM categories WHERE slug = $1 AND id != $2", *updateData.Slug, categoryID).Scan(&existingSlugCount)

			if err != nil {

				l.ErrorF("Error checking slug: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check slug uniqueness"})
				return
			}
			if existingSlugCount > 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Slug already in use"})

				return
			}

		}
		if updateData.Description != nil {

			updateQuery += ", description = $4"
			args = append(args, *updateData.Description)
		}
		//  ... (Add other fields as needed)
		if updateData.IsActive != nil {
			updateQuery += ", is_active = $5"
			args = append(args, *updateData.IsActive)
		}

		updateQuery += " WHERE id = $6"
		args = append(args, categoryID)

		_, err = tx.ExecContext(ctx, updateQuery, args...)
		if err != nil {

			l.ErrorF("Error updating category: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category"})
			return
		}

		if err = tx.Commit(); err != nil { // Commit transaction
			l.ErrorF("Error committing transaction: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Category updated successfully"})

	}
}

func UpdateCategoryStatus(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		categoryIDStr := c.Param("id")
		categoryID, err := uuid.Parse(categoryIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
			return
		}

		var req struct {
			IsActive *bool `json:"is_active"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {

			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if req.IsActive == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "isActive field is required"})
			return
		}

		_, err = app.DB.ExecContext(c, "UPDATE categories SET is_active = $1, updated = $2 WHERE id = $3", req.IsActive, time.Now(), categoryID)
		if err != nil {
			l.ErrorF("Failed to update category status: %v", err) // Log the error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category status"})

			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Category status updated"})
	}
}

func DeleteCategory(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		categoryIDStr := c.Param("id")
		categoryID, err := uuid.Parse(categoryIDStr)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
			return
		}

		ctx := context.Background()

		tx, err := app.DB.BeginTx(ctx, nil) // Start transaction
		if err != nil {
			l.DebugF("Error starting transaction: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
			return
		}

		defer tx.Rollback()

		_, err = tx.ExecContext(ctx, "DELETE FROM product_categories WHERE category_id = $1", categoryID) // Delete related products from junction table first.
		if err != nil {

			l.ErrorF("Failed to delete category from junction table: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category from products"})
			return

		}

		res, err := tx.ExecContext(ctx, "DELETE FROM categories WHERE id = $1", categoryID)

		if err != nil {
			l.DebugF("Failed to delete category: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category"})

			return
		}

		if err = tx.Commit(); err != nil { // Commit transaction

			l.ErrorF("Error committing transaction: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}

		rowsAffected, err := res.RowsAffected()

		if rowsAffected == 0 {

			c.JSON(http.StatusNotFound, gin.H{"message": "Category not found"})
			return
		}
		if err != nil {

			l.DebugF("Error getting rows affected: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get rows affected"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Category has been deleted successfully!"})
	}

}

// only admin can do it now
func AddProductToCategory(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		productIdStr := c.Param("product_id")
		productId, err := uuid.Parse(productIdStr)
		if err != nil {
			l.DebugF("Invalid product ID")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
			return
		}

		var req struct {
			Categories []uuid.UUID `json:"categories" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			l.DebugF("Invalid request body")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Check if product exists
		var productID uuid.UUID
		err = app.DB.QueryRowContext(c, "SELECT id FROM products WHERE id = $1", productId).Scan(&productID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			} else {
				l.ErrorF("Failed to fetch product: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product"})
			}
			return
		}

		// Check if category exists
		var categoryID []uuid.UUID
		query := `
		select id from categories 
			where id = any($1::uuid[])
		`
		// l.DebugF("caregories, %#v", req.Categories)
		rows, err := app.DB.QueryContext(c, query, pq.Array(req.Categories))
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
			} else {
				l.ErrorF("Failed to fetch category: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch category"})
			}
			return
		}
		defer rows.Close()

		for rows.Next() {
			var id uuid.UUID
			if err := rows.Scan(&id); err != nil {
				l.ErrorF("Error scanning category: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
				return
			}

			categoryID = append(categoryID, id)
		}

		if len(categoryID) != len(req.Categories) {
			l.DebugF("Category not found")
			c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
			return
		}

		// // Check if product is already in category
		// var existingProductID uuid.UUID
		// err = app.DB.QueryRowContext(c, "SELECT product_id FROM product_categories WHERE product_id = $1", productId).Scan(&existingProductID)
		// if err == nil {
		// 	l.DebugF("Product already in category")
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Product already in category"})
		// 	return
		// } else if !errors.Is(err, sql.ErrNoRows) {
		// 	l.ErrorF("Failed to check if product is in category: %v", err)
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check if product is in category"})
		// 	return
		// }

		// Add product to category
		query = `
		WITH category_ids AS (
			SELECT unnest($1::uuid[]) AS category_id
		)
		INSERT INTO product_categories (product_id, category_id)
		SELECT $2, category_id
		FROM category_ids
		ON CONFLICT (product_id, category_id) DO NOTHING
		`
		_, err = app.DB.ExecContext(c, query, pq.Array(req.Categories), productId)
		if err != nil {
			l.ErrorF("Failed to add product to category: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add product to category"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Product added to category"})
	}
}

func RemoveProductFromCategory(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		productIdStr := c.Param("product_id")
		productId, err := uuid.Parse(productIdStr)
		if err != nil {
			l.DebugF("Invalid product ID")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
			return
		}

		categoryIDStr := c.Param("category_id")
		categoryID, err := uuid.Parse(categoryIDStr)
		if err != nil {
			l.DebugF("Invalid category ID")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
			return
		}

		_, err = app.DB.ExecContext(c, "DELETE FROM product_categories WHERE product_id = $1 AND category_id = $2", productId, categoryID)
		if err != nil {
			l.ErrorF("Failed to remove product from category: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove product from category"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Product removed from category"})
	}
}
