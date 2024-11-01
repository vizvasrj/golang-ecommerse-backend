package user

import (
	"math"
	"net/http"
	"regexp"
	"src/pkg/conf"
	"src/pkg/module/merchant"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SearchUsers(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if the user is authenticated and has the Admin role
		userRole := c.MustGet("role").(UserRole)
		if userRole != RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}

		search := c.Query("search")
		regex, err := regexp.Compile("(?i)" + search)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid search query"})
			return
		}

		filter := bson.M{
			"$or": []bson.M{
				{"firstName": bson.M{"$regex": regex}},
				{"lastName": bson.M{"$regex": regex}},
				{"email": bson.M{"$regex": regex}},
			},
		}

		cursor, err := app.UserCollection.Find(c, filter, options.Find().SetProjection(bson.M{"password": 0, "_id": 0}))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}
		defer cursor.Close(c)

		var users []User
		if err = cursor.All(c, &users); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding users"})
			return
		}

		var searchUsers []UserSearch
		// Populate merchant field
		for i, user := range users {
			searchUsers = append(searchUsers, UserSearch{
				ID:          user.ID,
				Email:       user.Email,
				PhoneNumber: user.PhoneNumber,
				FirstName:   user.FirstName,
				LastName:    user.LastName,
				Role:        user.Role,
				Provider:    user.Provider,
				Avatar:      user.Avatar,
				Created:     user.Created,
				Updated:     user.Updated,
			})
			if user.Merchant != primitive.NilObjectID {
				var merchant merchant.Merchant
				err := app.MerchantCollection.FindOne(c, bson.M{"_id": user.Merchant}).Decode(&merchant)
				if err == nil {
					searchUsers[i].Merchant = merchant
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"users": searchUsers,
		})
	}
}

func FetchUsers(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse query parameters for pagination
		page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
		if err != nil || page < 1 {
			page = 1
		}
		limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
		if err != nil || limit < 1 {
			limit = 10
		}

		// Calculate skip value
		skip := (page - 1) * limit

		// Fetch users from the database with pagination
		findOptions := options.Find()
		findOptions.SetSort(bson.D{{Key: "created", Value: -1}})
		findOptions.SetLimit(int64(limit))
		findOptions.SetSkip(int64(skip))
		findOptions.SetProjection(bson.M{"password": 0, "_id": 0, "googleId": 0})

		cursor, err := app.UserCollection.Find(c, bson.M{}, findOptions)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}
		defer cursor.Close(c)

		var users []User
		if err = cursor.All(c, &users); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding users"})
			return
		}

		// Populate merchant field
		var searchUsers []UserSearch
		for i, user := range users {
			searchUsers = append(searchUsers, UserSearch{
				ID:          user.ID,
				Email:       user.Email,
				PhoneNumber: user.PhoneNumber,
				FirstName:   user.FirstName,
				LastName:    user.LastName,
				Role:        user.Role,
				Provider:    user.Provider,
				Avatar:      user.Avatar,
				Created:     user.Created,
				Updated:     user.Updated,
			})
			if user.Merchant != primitive.NilObjectID {
				var merchant merchant.Merchant
				err := app.MerchantCollection.FindOne(c, bson.M{"_id": user.Merchant}).Decode(&merchant)
				if err == nil {
					searchUsers[i].Merchant = merchant
				}
			}
		}

		// Get the total count of users
		count, err := app.UserCollection.CountDocuments(c, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error counting users"})
			return
		}

		// Return the users along with pagination details
		c.JSON(http.StatusOK, gin.H{
			"users":       searchUsers,
			"totalPages":  math.Ceil(float64(count) / float64(limit)),
			"currentPage": page,
			"count":       count,
		})
	}
}

func GetCurrentUser(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract the user ID from the request context
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found in request context"})
			return
		}

		// Convert userID to ObjectID
		objID, err := primitive.ObjectIDFromHex(userID.(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		// Fetch the user document from the database by ID
		var user User
		err = app.UserCollection.FindOne(c, bson.M{"_id": objID}, options.FindOne().SetProjection(bson.M{"password": 0})).Decode(&user)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}

		// Populate the merchant and brand fields
		var SearchedUser UserSearch
		SearchedUser.ID = user.ID
		SearchedUser.Email = user.Email
		SearchedUser.PhoneNumber = user.PhoneNumber
		SearchedUser.FirstName = user.FirstName
		SearchedUser.LastName = user.LastName
		SearchedUser.Role = user.Role
		SearchedUser.Provider = user.Provider
		SearchedUser.Avatar = user.Avatar
		SearchedUser.Created = user.Created
		SearchedUser.Updated = user.Updated

		if user.Merchant != primitive.NilObjectID {
			var merchant merchant.Merchant
			err := app.MerchantCollection.FindOne(c, bson.M{"_id": user.Merchant}).Decode(&merchant)
			if err == nil {
				SearchedUser.Merchant = merchant
			}
		}

		// Return the user document in the response
		c.JSON(http.StatusOK, gin.H{"user": SearchedUser})
	}
}

func UpdateUserProfile(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract the user ID from the request context
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found in request context"})
			return
		}

		// Convert userID to ObjectID
		objID, err := primitive.ObjectIDFromHex(userID.(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		// Parse the update data from the request body
		var update UserUpdate
		if err := c.ShouldBindJSON(&update); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Update the user document in the database
		filter := bson.M{"_id": objID}
		exclude := bson.M{
			"password":   0,
			"googleId":   0,
			"facebookId": 0,
		}
		updateResult := app.UserCollection.FindOneAndUpdate(c, filter, bson.M{"$set": update}, options.FindOneAndUpdate().SetReturnDocument(options.After).SetProjection(exclude))
		if updateResult.Err() != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}

		// Decode the updated user document
		var updatedUser User
		if err := updateResult.Decode(&updatedUser); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding updated user"})
			return
		}

		// Return the updated user document in the response
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Your profile is successfully updated!",
			"user":    updatedUser,
		})
	}
}
