package user

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"src/common"
	"src/l"
	"src/pkg/conf"
)

// Model Structs

type UserSearch struct {
	common.User
	Merchant common.Merchant `json:"merchant"`
}

type UserUpdate struct {
	FirstName   *string `json:"firstName"`
	LastName    *string `json:"lastName"`
	PhoneNumber *string `json:"phoneNumber"`
	Avatar      *string `json:"avatar"`
}

func SearchUsers(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {

		userRoleStr := c.MustGet("role").(string)
		userRole := common.GetUserRole(userRoleStr)
		if userRole != common.RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}

		search := c.Query("search")
		search = strings.TrimSpace(search) // Remove leading/trailing spaces

		query := `
            SELECT u.id, u.email, u.phone_number, u.first_name, u.last_name, u.role, u.provider, u.avatar, u.created, u.updated,
                   m.id AS merchant_id, m.name AS merchant_name, m.email AS merchant_email, m.phone_number AS merchant_phone_number, m.brand_name, m.business, m.is_active AS merchant_is_active, m.status AS merchant_status, m.updated AS merchant_updated, m.created AS merchant_created
            FROM users u
            LEFT JOIN merchants m ON u.id = m.user_id
            WHERE 1=1`

		var args []interface{}
		if search != "" {
			query += ` AND (u.first_name ILIKE $1 OR u.last_name ILIKE $1 OR u.email ILIKE $1)`
			args = append(args, "%"+search+"%")
		}

		rows, err := app.DB.QueryContext(c, query, args...)

		if err != nil {
			l.DebugF("Error querying users: %v", err)                                        // Log the error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search users"}) // Generic message for security
			return

		}
		defer rows.Close()

		var searchUsers []UserSearch
		for rows.Next() {

			var u common.User
			var m common.Merchant
			err := rows.Scan(
				&u.ID, &u.Email, &u.PhoneNumber, &u.FirstName, &u.LastName, &u.Role, &u.Provider, &u.Avatar, &u.Created, &u.Updated,
				&m.ID, &m.Name, &m.Email, &m.PhoneNumber, &m.BrandName, &m.Business, &m.IsActive, &m.Status, &m.Updated, &m.Created,
			)
			if err != nil {

				l.DebugF("Error scanning users: %v", err)                                       // More specific error message
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"}) // Generic message for security

				return
			}

			searchUsers = append(searchUsers, UserSearch{User: u, Merchant: m})

		}

		c.JSON(http.StatusOK, gin.H{"users": searchUsers})
	}
}

func FetchUsers(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

		if page < 1 {
			page = 1
		}
		if limit < 1 {

			limit = 10 // Or set to a reasonable default limit
		}

		offset := (page - 1) * limit

		rows, err := app.DB.QueryContext(c, `
			SELECT id, email, phone_number, first_name, last_name, role, provider, avatar, created, updated
			FROM users
            ORDER BY created DESC  -- Sort by the created timestamp in descending order
			LIMIT $1 OFFSET $2
		`, limit, offset) // Pass limit and offset as query parameters
		if err != nil {

			l.DebugF("Error fetching users: %v", err) // Log the error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
			return

		}
		defer rows.Close()

		users := []common.User{}
		for rows.Next() {

			var user common.User

			err := rows.Scan(&user.ID, &user.Email, &user.PhoneNumber, &user.FirstName, &user.LastName, &user.Role, &user.Provider, &user.Avatar, &user.Created, &user.Updated)

			if err != nil {

				l.ErrorF("Failed to scan user row: %v", err)                                        // Log the error for debugging
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan users data"}) // Generic message for security

				return

			}

			users = append(users, user)
		}

		// Fetch total users separately

		row := app.DB.QueryRow("SELECT COUNT(*) FROM users")

		var totalCount int
		err = row.Scan(&totalCount)

		if err != nil {
			l.DebugF("Error scanning users count: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan user count"})
			return
		}

		totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))

		c.JSON(http.StatusOK, gin.H{
			"users":        users,
			"total_pages":  totalPages,
			"current_page": page,
			"total_count":  totalCount, // Include the total count of users
		})

	}

}

func GetCurrentUser(app *conf.Config) gin.HandlerFunc { // Updated GetCurrentUser function
	return func(c *gin.Context) {
		userIDStr := c.GetString("userID")

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		var user common.User
		err = app.DB.QueryRowContext(c, "SELECT * FROM users WHERE id = $1", userID).Scan(&user.ID, &user.Email, &user.PhoneNumber, &user.FirstName, &user.LastName, &user.Password, &user.Provider, &user.GoogleID, &user.FacebookID, &user.Avatar, &user.Role, &user.ResetPasswordToken, &user.ResetPasswordExpires, &user.Updated, &user.Created)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) { // Check if it's a "no rows" error
				c.JSON(http.StatusNotFound, gin.H{"message": "User not found"})
			} else {

				l.ErrorF("Error fetching user : %v", err)

				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user data"}) // Generic message
			}

			return
		}

		var merchant common.Merchant
		err = app.DB.QueryRowContext(
			c, `SELECT id, user_id, name, email, phone_number, brand_name, business, is_active, status, updated, created 
			 FROM merchants WHERE user_id = $1`, user.ID,
		).Scan(&merchant.ID, &merchant.UserID, &merchant.Name, &merchant.Email, &merchant.PhoneNumber, &merchant.BrandName, &merchant.Business, &merchant.IsActive, &merchant.Status, &merchant.Updated, &merchant.Created) // Fetch all merchant details

		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			l.ErrorF("Error fetching merchant: %v", err) // Log the error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve merchant details"})
			return

		}

		userSearch := UserSearch{User: user, Merchant: merchant}

		c.JSON(http.StatusOK, gin.H{"user": userSearch})

	}
}

// UpdateUserProfile (updated below)

func UpdateUserProfile(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.GetString("userID")

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		var updateData UserUpdate
		if err := c.ShouldBindJSON(&updateData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		updateQuery := "UPDATE users SET updated = $1"
		args := []interface{}{time.Now()}
		argIndex := 2

		if updateData.FirstName != nil {
			updateQuery += fmt.Sprintf(", first_name = $%d", argIndex)
			args = append(args, *updateData.FirstName)
			argIndex++
		}
		if updateData.LastName != nil {
			updateQuery += fmt.Sprintf(", last_name = $%d", argIndex)
			args = append(args, *updateData.LastName)
			argIndex++
		}

		if updateData.PhoneNumber != nil {
			updateQuery += fmt.Sprintf(", phone_number = $%d", argIndex)
			args = append(args, *updateData.PhoneNumber)
			argIndex++
		}

		if updateData.Avatar != nil {
			updateQuery += fmt.Sprintf(", avatar = $%d", argIndex)
			args = append(args, *updateData.Avatar)
			argIndex++
		}

		updateQuery += fmt.Sprintf(" WHERE id = $%d", argIndex)
		args = append(args, userID)

		_, err = app.DB.ExecContext(c, updateQuery, args...)
		if err != nil {

			l.DebugF("Error updating user profile: %v", err) // Detailed log message
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
			return

		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Profile updated successfully"})
	}

}
