package auth

import (
	"encoding/json"
	"net/http"
	"net/url"
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
	"golang.org/x/oauth2"
)

func GoogleLogin(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		url := oauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
		c.Redirect(http.StatusTemporaryRedirect, url)
	}
}

func exchangeCodeForToken(code string) (string, error) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURI := os.Getenv("GOOGLE_REDIRECT_URI")
	grantType := "authorization_code"

	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("redirect_uri", redirectURI)
	data.Set("grant_type", grantType)

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", data)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	tokens := make(map[string]interface{})
	err = json.NewDecoder(resp.Body).Decode(&tokens)
	if err != nil {
		return "", err
	}
	l.DebugF("token came from google: %#v", tokens)

	return tokens["token"].(string), nil
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

type GoogleCode struct {
	Code string `json:"code"`
}

func GoogleCallbackPOST(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		l.InfoF("i get it here.\n")
		var codeDoc GoogleCode

		if err := c.ShouldBindJSON(&codeDoc); err != nil {
			l.ErrorF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		t, err := exchangeCodeForToken(codeDoc.Code)
		if err != nil {
			l.ErrorF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			// c.Redirect(http.StatusTemporaryRedirect, app.Env.ClientURL+"/login")
			return
		}
		l.DebugF("Code: %#v\n", codeDoc.Code)
		l.DebugF("Token %s", t)
		token, err := oauthConfig.Exchange(c, codeDoc.Code)
		if err != nil {
			l.ErrorF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			// c.Redirect(http.StatusTemporaryRedirect, app.Env.ClientURL+"/login")
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
		l.DebugF("User info: %#v\n", userInfo)

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
		// c.Redirect(http.StatusTemporaryRedirect, app.Env.ClientURL+"/auth/success?token="+jwtToken)
		c.JSON(http.StatusOK, gin.H{
			"token":    jwtToken,
			"redirect": app.Env.ClientURL + "/auth/success?token=" + jwtToken,
		})
	}
}
