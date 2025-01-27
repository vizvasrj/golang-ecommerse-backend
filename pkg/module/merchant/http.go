package merchant

import (
	"context"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"src/common"
	"src/l"
	"src/pkg/conf"
)

type MerchantAdd struct {
	Name        string `json:"name" binding:"required"`
	Email       string `json:"email" binding:"required"`
	PhoneNumber string `json:"phoneNumber" binding:"required"`
	BrandName   string `json:"brandName" binding:"required"`
	Business    string `json:"business" binding:"required"`
}

type MerchantUpdate struct {
	Name        *string    `json:"name"`
	Email       *string    `json:"email"`
	PhoneNumber *string    `json:"phoneNumber"`
	BrandName   *string    `json:"brandName"`
	Business    *string    `json:"business"`
	IsActive    *bool      `json:"isActive"`
	BrandID     *uuid.UUID `json:"brandId"`
	Status      *string    `json:"status"`
}

// HTTP Handlers (using database/sql and pq)

func AddMerchant(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var addMerchant MerchantAdd

		if err := c.ShouldBindJSON(&addMerchant); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input. Please provide all required fields."})
			return
		}

		userIDStr := c.GetString("userID") // Assuming you're using a middleware to set this
		userID, _ := uuid.Parse(userIDStr)

		// Check if email is already in use (using a single query)

		var emailCount int
		err := app.DB.QueryRowContext(c, "SELECT COUNT(*) FROM merchants WHERE email = $1", addMerchant.Email).Scan(&emailCount)

		if err != nil {

			l.DebugF("Error checking email uniqueness: %v", err) // Log the error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check email uniqueness"})
			return

		}

		if emailCount > 0 {

			c.JSON(http.StatusBadRequest, gin.H{"error": "Email address is already in use"})
			return
		}

		newMerchantID := uuid.New()

		_, err = app.DB.ExecContext(c, `
            INSERT INTO merchants (id, name, email, phone_number, brand_name, business, is_active, status, updated, created, user_id)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        `, newMerchantID, addMerchant.Name, addMerchant.Email, addMerchant.PhoneNumber, addMerchant.BrandName, addMerchant.Business, true, "Waiting Approval", time.Now(), time.Now(), userID) // Start with null user_id, setting valid to false

		if err != nil {

			l.DebugF("Error inserting merchant: %v", err)                                    // Log the error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add merchant"}) // Generic error message to the client
			return
		}

		// ... (send confirmation email logic - implement outside this function after successful insert)
		c.JSON(http.StatusOK, gin.H{
			"success":     true,
			"message":     "Merchant application submitted successfully!", // More appropriate message
			"merchant_id": newMerchantID.String(),                         // Return ID in string format
		})
	}

}

func SearchMerchants(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		searchQuery := c.Query("search")

		if searchQuery == "" {

			c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
			return
		}

		// Construct the SQL query (use ILIKE for case-insensitive matching)
		query := `
			SELECT id, user_id, name, email, phone_number, brand_name, business, is_active, status , created, updated
			FROM merchants 
			WHERE name ILIKE $1 OR email ILIKE $1 OR phone_number ILIKE $1 OR brand_name ILIKE $1 OR status::text ILIKE $1
		`
		// Execute query with database/sql
		rows, err := app.DB.QueryContext(c, query, "%"+searchQuery+"%") // Use parameterized query for security

		if err != nil {

			l.DebugF("Error searching merchants: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to perform search"}) // Don't reveal query details in error
			return
		}

		defer rows.Close()

		merchants := []common.Merchant{} // Initialize as empty slice to avoid null in response
		for rows.Next() {

			var merchant common.Merchant

			err := rows.Scan(

				&merchant.ID, &merchant.UserID, &merchant.Name, &merchant.Email, &merchant.PhoneNumber, &merchant.BrandName, &merchant.Business, &merchant.IsActive, &merchant.Status, &merchant.Created, &merchant.Updated)
			if err != nil {

				l.DebugF("Error scanning merchants: %v", err)                                          // Log the error
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan merchant data"}) // Generic error message
				return

			}
			merchants = append(merchants, merchant) // Append data to merchants slice
		}

		c.JSON(http.StatusOK, gin.H{"merchants": merchants}) // Return the results, even if empty

	}
}

func FetchAllMerchants(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get page and limit from query parameters
		page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
		if err != nil || page < 1 {
			page = 1
		}

		limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
		if err != nil || limit < 1 {
			limit = 10
		}

		offset := (page - 1) * limit

		rows, err := app.DB.QueryContext(c, `
            SELECT id, user_id, name, email, phone_number, business, is_active, status, updated, created
            FROM merchants
            ORDER BY created DESC
            LIMIT $1 OFFSET $2
        `, limit, offset)

		if err != nil {
			l.DebugF("Database query error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve merchants"})
			return
		}

		defer rows.Close()
		merchants := []common.Merchant{}

		for rows.Next() {
			var merchant common.Merchant
			err := rows.Scan(&merchant.ID, &merchant.UserID, &merchant.Name, &merchant.Email, &merchant.PhoneNumber, &merchant.Business, &merchant.IsActive, &merchant.Status, &merchant.Updated, &merchant.Created)
			if err != nil {
				l.DebugF("Error scanning merchant: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve merchants"})
				return
			}
			merchants = append(merchants, merchant)
		}

		var totalMerchants int
		err = app.DB.QueryRow("SELECT COUNT(*) FROM merchants").Scan(&totalMerchants)
		if err != nil {
			l.ErrorF("Error counting merchants: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count merchants"})
			return
		}

		totalPages := int(math.Ceil(float64(totalMerchants) / float64(limit)))

		c.JSON(http.StatusOK, gin.H{
			"merchants":       merchants,
			"total_pages":     totalPages,
			"current_page":    page,
			"total_merchants": totalMerchants,
		})
	}
}

func DisableMerchantAccount(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {

		authUserIDStr := c.GetString("userID") // Assuming you're using a middleware to set this
		authUserID, err := uuid.Parse(authUserIDStr)

		if err != nil {

			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return

		}

		userRoleStr := c.MustGet("role").(string)
		userRole := common.GetUserRole(userRoleStr)

		merchantIDStr := c.Param("id")
		merchantID, err := uuid.Parse(merchantIDStr)

		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid merchant ID"})

			return
		}

		var update struct {
			IsActive *bool `json:"isActive" binding:"required"` // Changed name for Postgres convention.
		}

		if err := c.ShouldBindJSON(&update); err != nil {

			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Authorization Check. Put these checks first for clarity and early exit
		if userRole != common.RoleAdmin { // Admins can disable any account
			if authUserID != merchantID { // Non-admins can only disable their own account
				c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
				return
			}
		}

		ctx := context.Background()
		l.DebugF("update.IsActive: %v", update.IsActive)

		_, err = app.DB.ExecContext(ctx, `UPDATE merchants SET is_active = $1, updated = $2 WHERE id = $3`, update.IsActive, time.Now(), merchantID)

		if err != nil {

			l.DebugF("Error updating merchant status: %v", err)                                        // Log the error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update merchant status"}) // Generic error message
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Merchant status updated successfully."})
	}

}

func ApproveMerchant(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		merchantIDStr := c.Param("id")
		merchantID, err := uuid.Parse(merchantIDStr)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid merchant ID"})
			return
		}

		// Update merchant status to "Approved" and set is_active to true
		ctx := context.Background()
		_, err = app.DB.ExecContext(ctx, `
            UPDATE merchants 
            SET status = $1, is_active = $2, updated = $3 
            WHERE id = $4
        `, common.Approved, true, time.Now(), merchantID)

		if err != nil {
			l.DebugF("Error approving merchant: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve merchant"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Merchant approved successfully."})
	}
}
