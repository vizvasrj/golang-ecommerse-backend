package address

import (
	"net/http"
	"src/pkg/conf"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func AddAddress(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var address Address
		if err := c.ShouldBindJSON(&address); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		userID, _ := c.Get("userID")
		address.User = userID.(primitive.ObjectID)
		address.Created = time.Now()
		address.Updated = time.Now()

		_, err := app.AddressCollection.InsertOne(c, address)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not add address"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Address has been added successfully!", "address": address})
	}
}

func GetAddresses(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")

		cursor, err := app.AddressCollection.Find(c, bson.M{"user": userID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch addresses"})
			return
		}
		defer cursor.Close(c)

		var addresses []Address
		if err = cursor.All(c, &addresses); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch addresses"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"addresses": addresses})
	}
}

func GetAddress(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		addressID, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address ID"})
			return
		}

		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var address Address
		err = app.AddressCollection.FindOne(c, bson.M{"_id": addressID, "user": userID}).Decode(&address)
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"message": "Cannot find Address with the id: " + c.Param("id")})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch address"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"address": address})
	}
}

func UpdateAddress(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		addressID, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address ID"})
			return
		}

		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		var update AddressUpdate
		if err := c.ShouldBindJSON(&update); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		filter := bson.M{"_id": addressID, "user": userID}
		updateResult, err := app.AddressCollection.UpdateOne(c, filter, bson.M{"$set": update})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update address"})
			return
		}

		if updateResult.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Address not found or does not belong to the user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Address has been updated successfully!"})
	}
}

func DeleteAddress(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		addressID, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address ID"})
			return
		}

		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		filter := bson.M{"_id": addressID, "user": userID}
		deleteResult, err := app.AddressCollection.DeleteOne(c, filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete address"})
			return
		}

		if deleteResult.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Address not found or does not belong to the user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Address has been deleted successfully!"})
	}
}

func SetDefaultAddress(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		addressID, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address ID"})
			return
		}

		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		_, err = app.AddressCollection.UpdateMany(c, bson.M{"user": userID}, bson.M{"$set": bson.M{"isDefault": false}})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update addresses"})
			return
		}

		var address Address
		err = app.AddressCollection.FindOneAndUpdate(c, bson.M{"_id": addressID, "user": userID}, bson.M{"$set": bson.M{"isDefault": true}}, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&address)
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"message": "Cannot find Address with the id: " + c.Param("id")})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update address"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Address has been set as default successfully!", "address": address})
	}
}
