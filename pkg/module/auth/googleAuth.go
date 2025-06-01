package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"src/common"
	"src/l"
	"src/pkg/conf"
	"src/pkg/middleware"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Model Structs
type InsertUserFromGmail struct {
	Email         string `json:"email"`
	FamilyName    string `json:"family_name"`
	GivenName     string `json:"given_name"`
	GoogleID      string `json:"_id"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	VerifiedEmail bool   `json:"verified_email"`
}

var googleOAuthConfig = &oauth2.Config{
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	RedirectURL:  "postmessage", // Or your redirect URL
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
	Endpoint:     google.Endpoint,
}

type GoogleCode struct {
	Code string `json:"code"`
}

// HTTP Handlers

func GoogleLogin(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		url := googleOAuthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
		c.Redirect(http.StatusTemporaryRedirect, url)
	}
}

func GoogleCallback(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Query("code")

		token, err := googleOAuthConfig.Exchange(c, code)
		if err != nil {
			l.ErrorF("OAuth exchange error: %v", err)
			c.Redirect(http.StatusTemporaryRedirect, app.Env.ClientURL+"/login?error=oauth_exchange")
			return
		}

		client := googleOAuthConfig.Client(c, token)
		resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
		if err != nil {
			l.ErrorF("Error getting user info: %v", err)
			c.Redirect(http.StatusTemporaryRedirect, app.Env.ClientURL+"/login?error=user_info")
			return
		}
		defer resp.Body.Close()

		var userInfo InsertUserFromGmail
		if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
			l.ErrorF("Decoding user info error: %v", err)
			c.Redirect(http.StatusTemporaryRedirect, app.Env.ClientURL+"/login?error=decode")
			return
		}

		ctx := context.Background()
		tx, err := app.DB.BeginTx(ctx, nil)
		if err != nil {
			l.ErrorF("Error beginning transaction: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
			return
		}
		defer tx.Rollback()

		var dbUser common.User
		err = tx.QueryRowContext(ctx, "SELECT id, email, first_name, last_name, role, provider FROM users WHERE email = $1", userInfo.Email).Scan(&dbUser.ID, &dbUser.Email, &dbUser.FirstName, &dbUser.LastName, &dbUser.Role, &dbUser.Provider)

		if err != nil {
			if err == sql.ErrNoRows {
				// User not found, create a new user
				dbUser.ID = uuid.New()
				dbUser.Email = userInfo.Email
				dbUser.FirstName = userInfo.GivenName
				dbUser.LastName = userInfo.FamilyName
				dbUser.Role = string(common.RoleMember) // Store as string
				dbUser.Provider = common.EmailProviderGoogle

				_, err = tx.ExecContext(ctx, `
					INSERT INTO users (id, email, first_name, last_name, role, provider, created, updated)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				`, dbUser.ID, dbUser.Email, dbUser.FirstName, dbUser.LastName, dbUser.Role, dbUser.Provider, time.Now(), time.Now())

				if err != nil {

					l.ErrorF("Error creating user: %v", err)
					c.Redirect(http.StatusTemporaryRedirect, app.Env.ClientURL+"/login?error=create_user")
					return
				}
			} else {

				l.ErrorF("Error querying user: %v", err)
				c.Redirect(http.StatusTemporaryRedirect, app.Env.ClientURL+"/login?error=query_user")
				return
			}
		}

		var merchantID string
		err = tx.QueryRowContext(ctx, "SELECT id FROM merchants WHERE user_id = $1", dbUser.ID).Scan(&merchantID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {

			l.ErrorF("Error finding merchant: %v", err)
			c.Redirect(http.StatusTemporaryRedirect, app.Env.ClientURL+"/login?error=find_merchant")
			return
		}

		if err := tx.Commit(); err != nil {
			l.ErrorF("Error committing transaction: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}

		sData := middleware.SignedDetails{
			Uid:        dbUser.ID.String(),
			Email:      dbUser.Email,
			FirstName:  dbUser.FirstName,
			LastName:   dbUser.LastName,
			Role:       dbUser.Role,
			MerchantID: merchantID,
		}

		tokenString, _, err := middleware.GenerateTokens(app, sData)
		if err != nil {
			l.ErrorF("Error generating tokens: %v", err)
			c.Redirect(http.StatusTemporaryRedirect, app.Env.ClientURL+"/login?error=generate_token")
			return
		}

		jwtToken := "Bearer " + tokenString
		c.Redirect(http.StatusTemporaryRedirect, app.Env.ClientURL+"/auth/success?token="+jwtToken)
	}
}

// GoogleCallbackPOST - (Implementation below)

func GoogleCallbackPOST(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var codeDoc GoogleCode

		if err := c.ShouldBindJSON(&codeDoc); err != nil {
			l.ErrorF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		token, err := googleOAuthConfig.Exchange(c, codeDoc.Code) // Exchange code for token
		if err != nil {
			l.ErrorF("OAuth exchange error: %v", err)
			c.Redirect(http.StatusTemporaryRedirect, app.Env.ClientURL+"/login?error=oauth_exchange")
			return
		}

		client := googleOAuthConfig.Client(c, token)
		resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
		if err != nil {
			l.ErrorF("Error getting user info: %v", err)
			c.Redirect(http.StatusTemporaryRedirect, app.Env.ClientURL+"/login?error=user_info")
			return
		}
		defer resp.Body.Close()

		var userInfo InsertUserFromGmail
		if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
			l.ErrorF("Decoding user info error: %v", err)
			c.Redirect(http.StatusTemporaryRedirect, app.Env.ClientURL+"/login?error=decode")
			return
		}

		ctx := context.Background()
		tx, err := app.DB.BeginTx(ctx, nil)
		if err != nil {
			l.ErrorF("Error beginning transaction: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
			return
		}
		defer tx.Rollback()

		var dbUser common.User
		err = tx.QueryRowContext(ctx, "SELECT id, email, first_name, last_name, role, provider FROM users WHERE email = $1", userInfo.Email).Scan(&dbUser.ID, &dbUser.Email, &dbUser.FirstName, &dbUser.LastName, &dbUser.Role, &dbUser.Provider)

		if err != nil {
			if err == sql.ErrNoRows {
				// User not found, create a new user
				dbUser.ID = uuid.New()
				dbUser.Email = userInfo.Email
				dbUser.FirstName = userInfo.GivenName
				dbUser.LastName = userInfo.FamilyName
				dbUser.Role = string(common.RoleMember) // Store as string
				dbUser.Provider = common.EmailProviderGoogle

				_, err = tx.ExecContext(ctx, `
					INSERT INTO users (id, email, first_name, last_name, role, provider, created, updated)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
				`, dbUser.ID, dbUser.Email, dbUser.FirstName, dbUser.LastName, dbUser.Role, dbUser.Provider, time.Now(), time.Now())

				if err != nil {

					l.ErrorF("Error creating user: %v", err)
					c.Redirect(http.StatusTemporaryRedirect, app.Env.ClientURL+"/login?error=create_user")
					return
				}
			} else {

				l.ErrorF("Error querying user: %v", err)
				c.Redirect(http.StatusTemporaryRedirect, app.Env.ClientURL+"/login?error=query_user")
				return
			}
		}

		var merchantID string
		err = tx.QueryRowContext(ctx, "SELECT id FROM merchants WHERE user_id = $1", dbUser.ID).Scan(&merchantID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {

			l.ErrorF("Error finding merchant: %v", err)
			c.Redirect(http.StatusTemporaryRedirect, app.Env.ClientURL+"/login?error=find_merchant")
			return
		}

		if err := tx.Commit(); err != nil {
			l.ErrorF("Error committing transaction: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}

		sData := middleware.SignedDetails{
			Uid:        dbUser.ID.String(),
			Email:      dbUser.Email,
			FirstName:  dbUser.FirstName,
			LastName:   dbUser.LastName,
			Role:       dbUser.Role,
			MerchantID: merchantID,
		}

		tokenString, _, err := middleware.GenerateTokens(app, sData)
		if err != nil {
			l.ErrorF("Error generating tokens: %v", err)
			c.Redirect(http.StatusTemporaryRedirect, app.Env.ClientURL+"/login?error=generate_token")
			return
		}

		jwtToken := "Bearer " + tokenString
		c.JSON(http.StatusOK, gin.H{
			"token":    jwtToken,
			"redirect": app.Env.ClientURL + "/auth/success?token=" + jwtToken,
		})
	}

}
