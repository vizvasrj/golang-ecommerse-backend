package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"src/common"
	"src/l"
	"src/pkg/conf"
	"src/pkg/middleware"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/guregu/null/v5"
	"golang.org/x/crypto/bcrypt"
)

// Request Structs

type UserRegister struct {
	Email     string `json:"email" binding:"required"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Password  string `json:"password" binding:"required"`
}

func Login(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			l.DebugF("Error binding JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		if req.Email == "" || req.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email and password are required"})
			return
		}

		var loggedInUser common.User
		err := app.DB.QueryRowContext(c, "SELECT id, email, password, first_name, last_name, role FROM users WHERE email = $1", req.Email).Scan(&loggedInUser.ID, &loggedInUser.Email, &loggedInUser.Password, &loggedInUser.FirstName, &loggedInUser.LastName, &loggedInUser.Role)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"}) // Don't reveal email existence
			} else {
				l.DebugF("Database query error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
			}
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(loggedInUser.Password), []byte(req.Password))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		merchantID := ""
		err = app.DB.QueryRow("SELECT id FROM merchants WHERE user_id = $1", loggedInUser.ID).Scan(&merchantID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) { // Handle case where merchant might not exist
			l.DebugF("Error checking for merchant: %s", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check merchant status"})
			return
		}

		sData := middleware.SignedDetails{
			Email:      loggedInUser.Email,
			FirstName:  loggedInUser.FirstName,
			LastName:   loggedInUser.LastName,
			Uid:        loggedInUser.ID.String(),
			Role:       loggedInUser.Role,
			MerchantID: merchantID, // Updated to pass merchantID directly
		}

		token, _, err := middleware.GenerateTokens(app, sData)

		if err != nil {
			l.DebugF("Error generating tokens: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"token":   "Bearer " + token,
			"user": gin.H{
				"id":        loggedInUser.ID,
				"firstName": loggedInUser.FirstName,
				"lastName":  loggedInUser.LastName,
				"email":     loggedInUser.Email,
				"role":      loggedInUser.Role,
			},
		})
	}
}

func Register(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UserRegister
		if err := c.ShouldBindJSON(&req); err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		if req.Email == "" || req.FirstName == "" || req.LastName == "" || req.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required"})
			return
		}

		var count int
		err := app.DB.QueryRowContext(c, "SELECT COUNT(*) FROM users WHERE email = $1", req.Email).Scan(&count)

		if err != nil {
			l.DebugF("Database error checking email: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check for existing user"})
			return
		}

		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email address is already in use"})
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			l.DebugF("Error hashing password: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		newUser := common.User{
			ID:        uuid.New(),
			Email:     req.Email,
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Password:  string(hashedPassword),
			Role:      "ROLE USER", // Set default role
			Created:   time.Now(),
			Updated:   null.TimeFrom(time.Now()),
		}

		_, err = app.DB.ExecContext(c, `
			INSERT INTO users (id, email, first_name, last_name, password, role, provider, created, updated)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, newUser.ID, newUser.Email, newUser.FirstName, newUser.LastName, newUser.Password, newUser.Role, "email", newUser.Created, newUser.Updated)

		if err != nil {
			l.DebugF("Error inserting user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		sData := middleware.SignedDetails{
			Email:      newUser.Email,
			FirstName:  newUser.FirstName,
			LastName:   newUser.LastName,
			Uid:        newUser.ID.String(),
			Role:       newUser.Role,
			MerchantID: "", //  no merchant ID during registration
		}

		token, _, err := middleware.GenerateTokens(app, sData)

		if err != nil {
			l.DebugF("Error generating token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"token":   "Bearer " + token,
			"user": gin.H{
				"id":        newUser.ID,
				"firstName": newUser.FirstName,
				"lastName":  newUser.LastName,
				"email":     newUser.Email,
				"role":      newUser.Role,
			},
		})
	}
}

func ForgotPassword(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email string `json:"email"`
		}

		if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email address"})
			return
		}

		var userID uuid.UUID // Changed type here
		err := app.DB.QueryRowContext(c, "SELECT id FROM users WHERE email = $1", req.Email).Scan(&userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{"error": "No user found for this email address"})
			} else {
				l.DebugF("Database query error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
			}
			return
		}

		resetToken := generateResetToken()
		expireTime := time.Now().Add(time.Hour)

		_, err = app.DB.ExecContext(c, `
            UPDATE users 
            SET reset_password_token = $1, reset_password_expires = $2
            WHERE id = $3
        `, resetToken, expireTime, userID)
		if err != nil {
			l.DebugF("Error updating reset token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}

		fmt.Printf("Reset token: %s, User ID: %s, Expires at: %v\n", resetToken, userID, expireTime)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Please check your email for the link to reset your password.",
		})
	}
}

func generateResetToken() string {
	buffer := make([]byte, 48)
	_, err := rand.Read(buffer)
	if err != nil {
		l.ErrorF("Failed to generate random token: %v", err) // Log the error
		return ""                                            // Or handle the error appropriately in your application
	}
	return hex.EncodeToString(buffer)
}

func ResetPasswordFromToken(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		resetToken := c.Param("token")
		var req struct {
			Password string `json:"password"`
		}

		if err := c.ShouldBindJSON(&req); err != nil || req.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request. Password is required."})
			return
		}

		var user struct {
			ID                   uuid.UUID `db:"id"`
			Email                string    `db:"email"`
			ResetPasswordExpires time.Time `db:"reset_password_expires"`
		}

		err := app.DB.QueryRowContext(c, `
			SELECT id, email, reset_password_expires 
			FROM users
			WHERE reset_password_token = $1
		`, resetToken).Scan(&user)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token."})
			} else {
				l.DebugF("Database error fetching user: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify token"})
			}
			return
		}

		if time.Now().After(user.ResetPasswordExpires) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Token has expired"})
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			l.DebugF("Password hashing error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		_, err = app.DB.ExecContext(c, `
			UPDATE users 
			SET password = $1, reset_password_token = NULL, reset_password_expires = NULL, updated = $2
			WHERE id = $3
		`, string(hashedPassword), time.Now(), user.ID)

		if err != nil {
			l.DebugF("Database error updating password: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Password reset successful"})
	}
}

func ResetPassword(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			CurrentPassword string `json:"currentPassword"`
			NewPassword     string `json:"newPassword"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		if req.CurrentPassword == "" || req.NewPassword == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Both current and new passwords are required"})
			return
		}

		userIDStr := c.GetString("userID")

		// Parse to UUID
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			l.DebugF("Error parsing userID %s : %v\n", userIDStr, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error parsing UUID"})
			return
		}

		var existingPassword string
		err = app.DB.QueryRowContext(c, `SELECT password FROM users WHERE id = $1`, userID).Scan(&existingPassword)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			} else {
				l.DebugF("Database error fetching user: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
			}
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(existingPassword), []byte(req.CurrentPassword))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect current password"})
			return
		}

		hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			l.DebugF("Error hashing new password: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		_, err = app.DB.ExecContext(c, `UPDATE users SET password = $1, updated = $2 WHERE id = $3`, string(hashedNewPassword), time.Now(), userID)
		if err != nil {
			l.DebugF("Database error updating password: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Password changed successfully"})
	}
}
