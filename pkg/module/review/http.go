package review

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"src/common"
	"src/l"
	"src/pkg/conf"
	"src/pkg/module/product"
)

// Model Structs

// GetAllReviews  (Updated below)
// GetProductReviewsBySlug (updated below)

func AddReview(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var reviewInput PutReviewInput
		if err := c.ShouldBindJSON(&reviewInput); err != nil {
			l.ErrorF("Error binding review input: %v", err) // Log the error
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		rating, err := strconv.ParseFloat(reviewInput.Rating, 64)
		if err != nil || rating < 1 || rating > 5 { // Validate rating
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rating. Must be between 1 and 5."})
			return
		}

		userIDStr := c.GetString("userID")
		userID, err := uuid.Parse(userIDStr) // Correctly parse the UUID

		if err != nil {
			l.DebugF("Invalid user ID: %v", err) // Log the error
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		// Product existence check (using QueryRowContext for efficiency)
		var productExists bool
		err = app.DB.QueryRowContext(c, "SELECT EXISTS (SELECT 1 FROM products WHERE id = $1)", reviewInput.ProductID).Scan(&productExists)

		if err != nil {
			l.ErrorF("Error checking for product: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify product"})
			return
		}
		if !productExists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"}) // Correct status code

			return

		}

		newReview := Review{

			ID:            uuid.New(),
			ProductID:     reviewInput.ProductID,
			UserID:        userID, // Correctly assign the UUID
			Title:         reviewInput.Title,
			Rating:        rating,
			Review:        reviewInput.Review,
			IsRecommended: reviewInput.IsRecommended,
			Status:        string(WaitingApproval), // Use the correct type
			Created:       time.Now(),
		}

		_, err = app.DB.ExecContext(c, `
			INSERT INTO reviews (id, product_id, user_id, title, rating, review, is_recommended, status, created)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, newReview.ID, newReview.ProductID, newReview.UserID, newReview.Title, newReview.Rating, newReview.Review, newReview.IsRecommended, newReview.Status, newReview.Created)

		if err != nil {
			l.ErrorF("Failed to insert review: %v", err)                                   // Log with more context
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add review"}) // More generic error message
			return

		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Your review has been added successfully and will appear when approved!",
			"review":  newReview,
		})
	}

}

func GetAllReviews(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {

		pageStr := c.DefaultQuery("page", "1")
		limitStr := c.DefaultQuery("limit", "10")

		page, err := strconv.Atoi(pageStr)

		if err != nil || page <= 0 {

			page = 1
		}

		limit, err := strconv.Atoi(limitStr)

		if err != nil || limit <= 0 {

			limit = 10 // Or another reasonable default
		}

		offset := (page - 1) * limit

		rows, err := app.DB.QueryContext(c, `
		SELECT r.id, r.product_id, r.user_id, r.title, r.rating, r.review, r.is_recommended, r.status, r.updated, r.created,
				u.first_name, u.last_name, u.email, p.name as product_name, p.slug as product_slug, p.image_url as product_image_url, p.price as product_price, p.taxable as product_taxable, p.is_active as product_is_active, p.brand_id as product_brand_id, p.merchant_id as product_merchant_id, p.updated as product_updated, p.created as product_created
		FROM reviews r
		JOIN products p ON r.product_id = p.id
		JOIN users u ON r.user_id = u.id
		ORDER BY r.created DESC                  -- Order by newest first
		LIMIT $1 OFFSET $2

		`, limit, offset)

		if err != nil {

			l.DebugF("Error fetching reviews: %v", err)                                        // Log the error for debugging
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reviews."}) // Don't leak query details
			return
		}

		defer rows.Close()

		var reviews []Review // Initialize an empty slice for reviews

		for rows.Next() {
			var review Review
			var user ReviewUser
			var productData product.Product

			err := rows.Scan(
				&review.ID, &review.ProductID, &review.UserID, &review.Title, &review.Rating, &review.Review, &review.IsRecommended, &review.Status, &review.Updated, &review.Created,
				&user.FirstName, &user.LastName, &user.Email, &productData.Name, &productData.Slug, &productData.ImageURL, &productData.Price, &productData.Taxable, &productData.IsActive, &productData.BrandID, &productData.MerchantID, &productData.Updated, &productData.Created,
			)
			if err != nil {
				l.ErrorF("Failed to scan review and user: %v", err) // Log the actual error for debugging
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan review data"})
				return
			}
			l.DebugF("Review: %v", review) // Log the review for debugging
			// Now associate the fetched user with the review.
			reviewUser := ReviewUser{
				ID:        review.UserID, // Associate user ID
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Email:     user.Email,
			}
			review.User = &reviewUser // Assuming you have a User field in your Review model

			// Now associate the fetched product with the review.
			productData.ID = review.ProductID
			review.Product = &productData // Assuming you have a Product field in your Review model

			reviews = append(reviews, review)

		}

		var totalReviews int

		err = app.DB.QueryRow("SELECT COUNT(*) FROM reviews").Scan(&totalReviews) // Count total reviews for pagination

		if err != nil {
			l.ErrorF("Error counting reviews: %v", err) // Log the error

			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch review count"})
			return

		}

		totalPages := int(math.Ceil(float64(totalReviews) / float64(limit))) // Calculate total pages

		c.JSON(http.StatusOK, gin.H{
			"reviews":       reviews,      // Return retrieved reviews
			"total_pages":   totalPages,   // Return total number of pages for pagination
			"current_page":  page,         // Return the current page number
			"total_reviews": totalReviews, // Return the total number of reviews

		})

	}
}

func GetProductReviewsBySlug(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {

		slug := c.Param("slug")
		var productID uuid.UUID
		err := app.DB.QueryRowContext(c, `SELECT id FROM products WHERE slug = $1`, slug).Scan(&productID)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{"message": "Product not found."})
			} else {

				l.ErrorF("Error retrieving product id %s : %v\n", slug, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product"})
			}

			return
		}
		l.DebugF("Product ID: %v", productID)
		rows, err := app.DB.QueryContext(c, `
            SELECT r.id, r.product_id, r.user_id, r.title, r.rating, r.review, r.is_recommended, r.status, r.updated, r.created,
                   u.first_name, u.last_name, u.email
            FROM reviews r
            JOIN users u ON r.user_id = u.id
            WHERE r.product_id = $1 AND r.status = $2  -- Get approved reviews
            ORDER BY r.created DESC                  -- Order by newest first
        `, productID, Approved) // Use Approved constant

		if err != nil {

			l.ErrorF("Error fetching reviews: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reviews"})
			return

		}
		defer rows.Close()

		reviews := make([]Review, 0)
		for rows.Next() {
			var review Review
			var user ReviewUser
			err := rows.Scan(
				&review.ID, &review.ProductID, &review.UserID, &review.Title, &review.Rating, &review.Review, &review.IsRecommended, &review.Status, &review.Updated, &review.Created,
				&user.FirstName, &user.LastName, &user.Email,
			)
			if err != nil {
				l.ErrorF("Failed to scan review and user: %v", err) // Log the actual error for debugging
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan review data"})
				return
			}
			l.DebugF("Review: %v", review) // Log the review for debugging
			// Now associate the fetched user with the review.
			reviewUser := ReviewUser{
				ID:        review.UserID, // Associate user ID
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Email:     user.Email,
			}
			review.User = &reviewUser // Assuming you have a User field in your Review model
			reviews = append(reviews, review)

		}
		c.JSON(http.StatusOK, gin.H{"reviews": reviews})
	}

}

func UpdateReview(app *conf.Config) gin.HandlerFunc { // Updated UpdateReview function

	return func(c *gin.Context) {
		userIDStr := c.GetString("userID")

		userID, err := uuid.Parse(userIDStr)
		if err != nil {

			l.DebugF("Invalid user ID: %v", err) // Log the specific error
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		reviewIDStr := c.Param("id")

		reviewID, err := uuid.Parse(reviewIDStr)

		if err != nil {
			l.DebugF("Invalid review ID: %v", err) // Log the error
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
			return
		}

		var updateReviewInput PutReviewInput
		if err := c.ShouldBindJSON(&updateReviewInput); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		rating, err := strconv.ParseFloat(updateReviewInput.Rating, 64)
		if err != nil || rating < 1 || rating > 5 { // Validate rating value
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rating. Must be 1-5"})
			return
		}

		ctx := context.Background()
		tx, err := app.DB.BeginTx(ctx, nil)
		if err != nil {
			l.ErrorF("Failed to begin transaction: %v", err) // More specific error message
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
			return
		}
		defer tx.Rollback() // Defer rollback

		// Check if the review exists and belongs to the user inside the transaction

		var reviewExists bool
		err = tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM reviews WHERE id = $1 AND user_id = $2)", reviewID, userID).Scan(&reviewExists)

		if err != nil {

			l.ErrorF("Error checking for review: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify review"})
			return

		}

		if !reviewExists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Review not found or you are not authorized to update it."})
			return
		}

		_, err = tx.ExecContext(ctx, `
			UPDATE reviews 
			SET title = $1, rating = $2, review = $3, is_recommended = $4, updated = $5
			WHERE id = $6 AND user_id = $7
		`, updateReviewInput.Title, rating, updateReviewInput.Review, updateReviewInput.IsRecommended, time.Now(), reviewID, userID)
		if err != nil {

			l.DebugF("Error updating review: %v", err)                                        // Detailed logging
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update review"}) // Generic message to client
			return
		}

		if err := tx.Commit(); err != nil { // Commit transaction

			l.DebugF("Transaction commit failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})

			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Review updated successfully!"})
	}

}

func ApproveReview(app *conf.Config) gin.HandlerFunc { // Updated ApproveReview
	return func(c *gin.Context) {
		reviewIDStr := c.Param("reviewId")

		reviewID, err := uuid.Parse(reviewIDStr)

		if err != nil {

			l.DebugF("Invalid review ID: %v", err) // Log the error
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
			return
		}

		merchantIDStr := c.GetString("merchantID") // From authentication middleware

		merchantID, err := uuid.Parse(merchantIDStr)
		if err != nil {

			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid merchant ID"})
			return
		}

		// Get the review and product details inside a transaction for consistency

		ctx := context.Background()
		tx, err := app.DB.BeginTx(ctx, nil)

		if err != nil {

			l.ErrorF("Error starting transaction : %v", err)                                      // Log the error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"}) // Generic error message
			return
		}
		defer tx.Rollback() // Important to defer the rollback

		var productID uuid.UUID

		err = tx.QueryRowContext(ctx, "SELECT product_id FROM reviews WHERE id = $1", reviewID).Scan(&productID)
		if err != nil {

			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
			} else {
				l.ErrorF("Error fetching review: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve review"})
			}
			return
		}

		var productMerchantID uuid.UUID
		err = tx.QueryRowContext(ctx, "SELECT merchant_id FROM products WHERE id = $1", productID).Scan(&productMerchantID)

		if err != nil {

			if errors.Is(err, sql.ErrNoRows) {

				c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			} else {

				l.DebugF("Error fetching product's merchant ID : %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product merchant ID"})

			}
			return

		}

		role := c.MustGet("role")
		roleType := common.GetUserRole(role)

		if productMerchantID != merchantID && roleType != common.RoleAdmin {

			c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to approve review"}) // Correct status code
			return

		}

		_, err = tx.ExecContext(ctx, "UPDATE reviews SET status = $1, updated = $2 WHERE id = $3", Approved, time.Now(), reviewID) // Use correct constant and time

		if err != nil {

			l.DebugF("Error approving review: %v", err) // More detailed message

			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve review"}) // Generic error message to the user

			return

		}

		if err = tx.Commit(); err != nil {

			l.DebugF("Transaction commit failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})

			return

		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Review approved"})
	}

}

func RejectReview(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		reviewIDStr := c.Param("reviewId")
		reviewID, err := uuid.Parse(reviewIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
			return
		}

		merchantIDStr := c.GetString("merchantID")
		merchantID, err := uuid.Parse(merchantIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid merchant ID"})
			return
		}

		ctx := context.Background()
		tx, err := app.DB.BeginTx(ctx, nil)
		if err != nil {
			l.ErrorF("Error starting transaction: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
			return
		}
		defer tx.Rollback()

		var productID uuid.UUID
		err = tx.QueryRowContext(ctx, "SELECT product_id FROM reviews WHERE id = $1", reviewID).Scan(&productID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
			} else {
				l.ErrorF("Error fetching review's product ID: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve review"})
			}
			return
		}

		var productMerchantID uuid.UUID
		err = tx.QueryRowContext(ctx, "SELECT merchant_id FROM products WHERE id = $1", productID).Scan(&productMerchantID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			} else {
				l.ErrorF("Error fetching product's merchant ID: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product merchant ID"})
			}
			return
		}

		role := c.MustGet("role")
		roleType := common.GetUserRole(role)

		if productMerchantID != merchantID && roleType != common.RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to reject review"})
			return
		}

		_, err = tx.ExecContext(ctx, "UPDATE reviews SET status = $1, updated = $2 WHERE id = $3", Rejected, time.Now(), reviewID)
		if err != nil {
			l.ErrorF("Error rejecting review: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject review"})
			return
		}

		if err = tx.Commit(); err != nil {
			l.ErrorF("Transaction commit failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Review rejected"})
	}
}

func DeleteReview(app *conf.Config) gin.HandlerFunc { // Updated DeleteReview function

	return func(c *gin.Context) {
		reviewIDStr := c.Param("id")

		reviewID, err := uuid.Parse(reviewIDStr)

		if err != nil {

			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
			return

		}

		userIDStr := c.GetString("userID")
		userID, err := uuid.Parse(userIDStr) // Parse the user ID from the claim

		if err != nil {
			l.DebugF("Error parsing user ID: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		userRoleStr := c.MustGet("role").(string)
		userRole := common.GetUserRole(userRoleStr)

		ctx := context.Background()
		tx, err := app.DB.BeginTx(ctx, nil) // Start transaction

		if err != nil {

			l.ErrorF("Error starting transaction: %v", err) // Log the error with more context

			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"}) // Generic message for security
			return

		}
		defer tx.Rollback() // Defer rollback

		var reviewUserID uuid.UUID

		err = tx.QueryRowContext(ctx, "SELECT user_id FROM reviews WHERE id = $1", reviewID).Scan(&reviewUserID)

		if err != nil {

			if errors.Is(err, sql.ErrNoRows) { // More specific error check

				c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
			} else {

				l.DebugF("Error fetching review's user ID: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch review"})

			}
			return

		}

		if userRole == common.RoleMerchant {
			// Merchant can delete reviews for products they own
			var productID uuid.UUID

			err = tx.QueryRowContext(ctx, "SELECT product_id FROM reviews WHERE id = $1", reviewID).Scan(&productID)
			if err != nil {
				l.ErrorF("Error fetching product ID from review: %v", err) // Detailed log message
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve product information"})

				return
			}

			var merchantID uuid.UUID
			err = tx.QueryRowContext(ctx, "SELECT merchant_id FROM products WHERE id = $1", productID).Scan(&merchantID)

			if err != nil {

				l.DebugF("Failed to retrieve merchant ID for product: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify product ownership"})

				return
			}

			// fetch user_id from merchants table using merchantID, then compare it with userID from token.

			var merchantUserID uuid.UUID
			err = tx.QueryRowContext(ctx, `select user_id from merchants where id = $1`, merchantID).Scan(&merchantUserID)

			if err != nil {
				l.DebugF("Failed to retrieve merchant user ID: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify product ownership"}) // Generic message

				return
			}

			if userID != merchantUserID {
				c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to delete this review"})
				return
			}

		} else if userRole != common.RoleAdmin && userID != reviewUserID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to delete this review"})
			return
		}

		_, err = tx.ExecContext(ctx, "DELETE FROM reviews WHERE id = $1", reviewID) // Use ExecContext

		if err != nil {
			l.DebugF("Error deleting review : %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete review"})
			return
		}

		if err := tx.Commit(); err != nil { // Commit transaction

			l.DebugF("Transaction commit failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})

			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Review deleted successfully"})
	}
}
