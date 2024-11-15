package user

import (
	"fmt"
	"net/http"
	"regexp"
	"src/common"
	"src/l"
	"src/pkg/conf"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func AddMerchant(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var addMerchant MerchantAdd

		if err := c.ShouldBindJSON(&addMerchant); err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "You must enter your name, email, business description, phone number, and email address."})
			return
		}

		// Check if a merchant with the given email already exists
		var existingMerchant Merchant
		err := app.MerchantCollection.FindOne(c, bson.M{"email": addMerchant.Email}).Decode(&existingMerchant)
		if err == nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "That email address is already in use."})
			return
		}

		// Save the merchant document to the database
		result, err := app.MerchantCollection.InsertOne(c, addMerchant)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}

		// todo Send a confirmation email
		// err = app.Mailgun.SendEmail(req.Email, "merchant-application")
		// if err != nil {
		l.DebugF("Error: %v", err)
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send confirmation email."})
		// 	return
		// }

		// Return the created merchant document in the response
		c.JSON(http.StatusOK, gin.H{
			"success":  true,
			"message":  "We received your request! We will reach you on your phone number " + addMerchant.PhoneNumber + "!",
			"merchant": result.InsertedID,
		})
	}
}

func SearchMerchants(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		search := c.Query("search")
		if search == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
			return
		}

		// Create a regular expression for the search term
		regex, err := regexp.Compile("(?i)" + search)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid search query"})
			return
		}

		// Query the database for merchants matching the search term
		filter := bson.M{
			"$or": []bson.M{
				{"phoneNumber": bson.M{"$regex": regex}},
				{"email": bson.M{"$regex": regex}},
				{"name": bson.M{"$regex": regex}},
				{"brandName": bson.M{"$regex": regex}},
				{"status": bson.M{"$regex": regex}},
			},
		}

		cursor, err := app.MerchantCollection.Find(c, filter, options.Find().SetProjection(bson.M{"brand": 1, "name": 1}))
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}
		defer cursor.Close(c)

		var merchants []Merchant
		if err := cursor.All(c, &merchants); err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding merchants"})
			return
		}

		// Return the matching merchants in the response
		c.JSON(http.StatusOK, gin.H{"merchants": merchants})
	}
}

func FetchAllMerchants(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		pageStr := c.DefaultQuery("page", "1")
		limitStr := c.DefaultQuery("limit", "10")

		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
			return
		}

		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit number"})
			return
		}

		// Calculate skip value
		skip := (page - 1) * limit

		// Query the database to fetch merchants with pagination
		findOptions := options.Find()
		findOptions.SetSort(bson.D{{Key: "created", Value: -1}})
		findOptions.SetLimit(int64(limit))
		findOptions.SetSkip(int64(skip))

		cursor, err := app.MerchantCollection.Find(c, bson.M{}, findOptions)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}
		defer cursor.Close(c)

		var merchants []Merchant
		if err := cursor.All(c, &merchants); err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding merchants"})
			return
		}

		// Count the total number of merchants
		count, err := app.MerchantCollection.CountDocuments(c, bson.M{})
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error counting merchants"})
			return
		}

		// Return the merchants along with pagination details in the response
		c.JSON(http.StatusOK, gin.H{
			"merchants":   merchants,
			"totalPages":  (count + int64(limit) - 1) / int64(limit), // Calculate total pages
			"currentPage": page,
			"count":       count,
		})
	}
}

func DisableMerchantAccount(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := c.MustGet("userID").(primitive.ObjectID)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		userRole, ok := c.MustGet("role").(common.UserRole)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		merchantId := c.Param("id")
		var update struct {
			IsActive bool `json:"isActive"`
		}
		if err := c.ShouldBindJSON(&update); err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		merchantObjectId, err := primitive.ObjectIDFromHex(merchantId)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid merchant ID"})
			return
		}

		// Get the authenticated user's ID and role
		// authUserId := c.GetString("userId")
		// authUserRole := c.GetString("userRole")

		// Check if the authenticated user is the same as the merchant or an admin
		if userRole != common.RoleAdmin || userID != merchantObjectId {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to perform this action"})
			return
		}

		query := bson.M{"_id": merchantObjectId}
		updateDoc := bson.M{"$set": bson.M{"isActive": update.IsActive}}

		var merchantDoc Merchant
		err = app.MerchantCollection.FindOneAndUpdate(c, query, updateDoc).Decode(&merchantDoc)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				l.DebugF("Error: %v", err)
				c.JSON(http.StatusNotFound, gin.H{"error": "Merchant not found"})
			} else {
				l.DebugF("Error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating merchant"})
			}
			return
		}

		if !update.IsActive {
			// todo unimplemented
			fmt.Println("unimplemented")
			// if err := deactivateBrand(merchantId); err != nil {
			// l.DebugF("Error: %v", err)
			// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deactivating brand"})
			// 	return
			// }
			// if err := mailgun.SendEmail(merchantDoc.Email, "merchant-deactivate-account"); err != nil {
			// l.DebugF("Error: %v", err)
			// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Error sending email"})
			// 	return
			// }
		}

		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}
