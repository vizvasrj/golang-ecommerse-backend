package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"src/common"
	"src/l"
	"src/pkg/conf"
	"src/pkg/middleware"
	"src/pkg/module/user"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// var secret = "your_jwt_secret"
// var tokenLife = time.Hour * 24
var oauthConfig = &oauth2.Config{
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	RedirectURL:  "https://backend.aapan.shop/api/auth/google/callback",
	// RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
	Scopes:   []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
	Endpoint: google.Endpoint,
}

func Login(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		if req.Email == "" || req.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email and password are required"})
			return
		}

		var logedInUser user.User
		err := app.UserCollection.FindOne(c, bson.M{"email": req.Email}).Decode(&logedInUser)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "No user found for this email address"})
			return
		}

		// if logedInUser.Provider != user.EmailProviderEmail {
		// 	l.DebugF("Error: %v", err)
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Email address is already in use with another provider"})
		// 	return
		// }

		err = bcrypt.CompareHashAndPassword([]byte(logedInUser.Password), []byte(req.Password))
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password incorrect"})
			return
		}
		sData := middleware.SignedDetails{
			Email:      logedInUser.Email,
			FirstName:  logedInUser.FirstName,
			LastName:   logedInUser.LastName,
			Uid:        logedInUser.ID.Hex(),
			Role:       logedInUser.Role,
			MerchantID: logedInUser.Merchant.Hex(),
		}
		token, _, err := middleware.GenerateTokens(app, sData)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"token":   "Bearer " + token,
			"user": gin.H{
				"id":        logedInUser.ID,
				"firstName": logedInUser.FirstName,
				"lastName":  logedInUser.LastName,
				"email":     logedInUser.Email,
				"role":      logedInUser.Role,
			},
		})
	}
}

func Register(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {

		var req struct {
			Email        string `json:"email"`
			FirstName    string `json:"firstName"`
			LastName     string `json:"lastName"`
			Password     string `json:"password"`
			IsSubscribed bool   `json:"isSubscribed"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		if req.Email == "" || req.FirstName == "" || req.LastName == "" || req.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "All fields are required"})
			return
		}

		var existingUser struct {
			ID string `bson:"_id"`
		}
		err := app.UserCollection.FindOne(c, bson.M{"email": req.Email}).Decode(&existingUser)
		if err == nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email address is already in use"})
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		updateUser := bson.M{
			"email":     req.Email,
			"password":  string(hash),
			"firstName": req.FirstName,
			"lastName":  req.LastName,
			"provider":  "email",
		}
		insertResult, err := app.UserCollection.InsertOne(c, updateUser)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		var returnedUser user.User
		err = app.UserCollection.FindOne(c, bson.M{"_id": insertResult.InsertedID}).Decode(&returnedUser)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch created user"})
			return
		}

		sData := middleware.SignedDetails{
			Email:      returnedUser.Email,
			FirstName:  returnedUser.FirstName,
			LastName:   returnedUser.LastName,
			Uid:        returnedUser.ID.Hex(),
			Role:       returnedUser.Role,
			MerchantID: returnedUser.Merchant.Hex(),
		}

		token, _, err := middleware.GenerateTokens(app, sData)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"token":   "Bearer " + token,
			"user": gin.H{
				"id":        returnedUser.ID,
				"firstName": returnedUser.FirstName,
				"lastName":  returnedUser.LastName,
				"email":     returnedUser.Email,
				"role":      returnedUser.Role,
			},
		})
	}
}

func ForgotPassword(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email string `json:"email"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "You must enter an email address."})
			return
		}

		// collection := getCollection("users")
		var user struct {
			Email                string `bson:"email"`
			ResetPasswordToken   string `bson:"resetPasswordToken"`
			ResetPasswordExpires int64  `bson:"resetPasswordExpires"`
		}
		err := app.UserCollection.FindOne(c, bson.M{"email": req.Email}).Decode(&user)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "No user found for this email address."})
			return
		}

		buffer := make([]byte, 48)
		_, err = rand.Read(buffer)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate reset token"})
			return
		}
		resetToken := hex.EncodeToString(buffer)
		expireTime := time.Now().Add(time.Hour).Unix()

		update := bson.M{
			"$set": bson.M{
				"resetPasswordToken":   resetToken,
				"resetPasswordExpires": expireTime,
			},
		}
		_, err = app.UserCollection.UpdateOne(c, bson.M{"email": req.Email}, update)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}

		fmt.Printf("Reset token: %s expire time: %v\n", resetToken, expireTime)

		// from := mail.NewEmail("Example User", "test@example.com")
		// subject := "Password Reset"
		// to := mail.NewEmail("Example User", user.Email)
		// plainTextContent := "Please use the following token to reset your password: " + resetToken
		// htmlContent := "<strong>Please use the following token to reset your password: " + resetToken + "</strong>"
		// message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
		// client := sendgrid.NewSendClient("YOUR_SENDGRID_API_KEY")
		// _, err = client.Send(message)
		// if err != nil {
		l.DebugF("Error: %v", err)
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email"})
		// 	return
		// }

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Please check your email for the link to reset your password.",
		})
	}
}

func ResetPasswordFromToken(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param("token")
		var req struct {
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "You must enter a password."})
			return
		}

		// collection := getCollection("users")
		var user struct {
			Email                string `bson:"email"`
			Password             string `bson:"password"`
			ResetPasswordToken   string `bson:"resetPasswordToken"`
			ResetPasswordExpires int64  `bson:"resetPasswordExpires"`
		}
		err := app.UserCollection.FindOne(c, bson.M{
			"resetPasswordToken":   token,
			"resetPasswordExpires": bson.M{"$gt": time.Now().Unix()},
		}).Decode(&user)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Your token has expired. Please attempt to reset your password again."})
			return
		}

		salt, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		update := bson.M{
			"$set": bson.M{
				"password":             string(salt),
				"resetPasswordToken":   nil,
				"resetPasswordExpires": nil,
			},
		}
		_, err = app.UserCollection.UpdateOne(c, bson.M{"email": user.Email}, update)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}

		// from := mail.NewEmail("Example User", "test@example.com")
		// subject := "Password Reset Confirmation"
		// to := mail.NewEmail("Example User", user.Email)
		// plainTextContent := "Your password has been successfully reset."
		// htmlContent := "<strong>Your password has been successfully reset.</strong>"
		// message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
		// client := sendgrid.NewSendClient("YOUR_SENDGRID_API_KEY")
		// _, err = client.Send(message)
		// if err != nil {
		l.DebugF("Error: %v", err)
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email"})
		// 	return
		// }

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Password changed successfully. Please login with your new password.",
		})
	}
}

func ResetPassword(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Password        string `json:"password"`
			ConfirmPassword string `json:"confirmPassword"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		email := c.GetString("email")
		if email == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthenticated"})
			return
		}

		if req.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You must enter a password."})
			return
		}

		// collection := getCollection("users")
		var user struct {
			Email    string `bson:"email"`
			Password string `bson:"password"`
		}
		err := app.UserCollection.FindOne(c, bson.M{"email": email}).Decode(&user)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "That email address is already in use."})
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Please enter your correct old password."})
			return
		}

		salt, err := bcrypt.GenerateFromPassword([]byte(req.ConfirmPassword), bcrypt.DefaultCost)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		update := bson.M{
			"$set": bson.M{
				"password": string(salt),
			},
		}
		_, err = app.UserCollection.UpdateOne(c, bson.M{"email": user.Email}, update)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}

		// from := mail.NewEmail("Example User", "test@example.com")
		// subject := "Password Reset Confirmation"
		// to := mail.NewEmail("Example User", user.Email)
		// plainTextContent := "Your password has been successfully reset."
		// htmlContent := "<strong>Your password has been successfully reset.</strong>"
		// message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
		// client := sendgrid.NewSendClient("YOUR_SENDGRID_API_KEY")
		// _, err = client.Send(message)
		// if err != nil {
		l.DebugF("Error: %v", err)
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send email"})
		// 	return
		// }

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Password changed successfully. Please login with your new password.",
		})
	}
}

func GoogleLogin(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		url := oauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
		c.Redirect(http.StatusTemporaryRedirect, url)
	}
}

func GoogleCallback(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		l.InfoF("i get it here.\n")
		code := c.Query("code")
		token, err := oauthConfig.Exchange(c, code)
		if err != nil {
			l.ErrorF("Error: %v", err)
			c.Redirect(http.StatusTemporaryRedirect, app.Env.ClientURL+"/login")
			return
		}

		client := oauthConfig.Client(c, token)
		resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
		if err != nil {
			l.ErrorF("Error: %v", err)
			c.Redirect(http.StatusTemporaryRedirect, app.Env.ClientURL+"/login")
			return
		}
		defer resp.Body.Close()

		// Parse user info from response
		// var userInfo InsertUserFromGmail	 {
		// 	ID    string `json:"id"`
		// 	Email string `json:"email"`
		// }
		var userInfo user.InsertUserFromGmail
		if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
			l.ErrorF("Error: %v", err)
			c.Redirect(http.StatusTemporaryRedirect, app.Env.ClientURL+"/login")
			return
		}
		// l.DebugF("User info: %#v\n", userInfo)

		// find user in db
		var dbUser user.User
		err = app.UserCollection.FindOne(c, bson.M{"email": userInfo.Email}).Decode(&dbUser)
		if err != nil {
			// create user

			insertUser := user.User{
				Email:     userInfo.Email,
				Provider:  user.EmailProviderGoogle,
				Created:   time.Now(),
				Updated:   time.Now(),
				FirstName: userInfo.GivenName,
				LastName:  userInfo.FamilyName,
				Role:      common.RoleMember,
			}
			insertResult, err := app.UserCollection.InsertOne(c, insertUser)
			if err != nil {
				l.ErrorF("Error: %v", err)
				c.Redirect(http.StatusTemporaryRedirect, app.Env.ClientURL+"/login")
				return
			}
			err = app.UserCollection.FindOne(c, bson.M{"_id": insertResult.InsertedID}).Decode(&dbUser)
			if err != nil {
				l.ErrorF("Error: %v", err)
				c.Redirect(http.StatusTemporaryRedirect, app.Env.ClientURL+"/login")
				return
			}

		}

		// Generate JWT token
		var merchantID string
		if dbUser.Merchant != primitive.NilObjectID {
			merchantID = dbUser.Merchant.Hex()
		}
		sData := middleware.SignedDetails{
			Uid:        dbUser.ID.Hex(),
			Email:      dbUser.Email,
			FirstName:  dbUser.FirstName,
			LastName:   dbUser.LastName,
			Role:       dbUser.Role,
			MerchantID: merchantID,
		}
		tokenString, _, err := middleware.GenerateTokens(app, sData)
		if err != nil {
			l.ErrorF("Error: %v", err)
			c.Redirect(http.StatusTemporaryRedirect, app.Env.ClientURL+"/login")
			return
		}

		jwtToken := "Bearer " + tokenString
		c.Redirect(http.StatusTemporaryRedirect, app.Env.ClientURL+"/auth/success?token="+jwtToken)
	}
}
