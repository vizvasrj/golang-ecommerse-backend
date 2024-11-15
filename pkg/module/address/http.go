package address

import (
	"net/http"
	"src/l"
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
		var address AddressAdd
		if err := c.ShouldBindJSON(&address); err != nil {

			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		userID, ok := c.MustGet("userID").(string)
		if !ok {
			l.DebugF("Error: %v", "User not authenticated")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}
		objectId, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}
		var addressAdd Address
		addressAdd.User = objectId
		addressAdd.Created = time.Now()
		addressAdd.Updated = time.Now()
		addressAdd.Address = address.Address
		addressAdd.City = address.City
		addressAdd.State = address.State
		addressAdd.Country = address.Country
		addressAdd.ZipCode = address.ZipCode
		addressAdd.IsDefault = address.IsDefault

		_, err = app.AddressCollection.InsertOne(c, addressAdd)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not add address"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Address has been added successfully!", "address": address})
	}
}

func GetAddresses(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {

		uid := c.MustGet("userID").(string)
		userID, err := primitive.ObjectIDFromHex(uid)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		cursor, err := app.AddressCollection.Find(c, bson.M{"user": userID})
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch addresses"})
			return
		}
		defer cursor.Close(c)

		var addresses []Address
		if err = cursor.All(c, &addresses); err != nil {
			l.DebugF("Error: %v", err)
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
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address ID"})
			return
		}

		userID, exists := c.Get("userID")
		if !exists {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		userObjectId, err := primitive.ObjectIDFromHex(userID.(string))
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		var address Address
		err = app.AddressCollection.FindOne(c, bson.M{"_id": addressID, "user": userObjectId}).Decode(&address)
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"message": "Cannot find Address with the id: " + c.Param("id")})
			return
		} else if err != nil {
			l.DebugF("Error: %v", err)
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
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address ID"})
			return
		}

		userID, exists := c.MustGet("userID").(string)
		if !exists {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		objectUserId, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		var update AddressAdd
		if err := c.ShouldBindJSON(&update); err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		filter := bson.M{"_id": addressID, "user": objectUserId}
		updateResult, err := app.AddressCollection.UpdateOne(c, filter, bson.M{"$set": update})
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update address"})
			return
		}

		if updateResult.MatchedCount == 0 {
			l.DebugF("Error: %v", err)
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
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address ID"})
			return
		}

		userID, exists := c.Get("userID")
		if !exists {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		filter := bson.M{"_id": addressID, "user": userID}
		deleteResult, err := app.AddressCollection.DeleteOne(c, filter)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete address"})
			return
		}

		if deleteResult.DeletedCount == 0 {
			l.DebugF("Error: %v", err)
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
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address ID"})
			return
		}

		userID, exists := c.Get("userID")
		if !exists {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}
		objectIdUserId, err := primitive.ObjectIDFromHex(userID.(string))
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		_, err = app.AddressCollection.UpdateMany(c, bson.M{"user": objectIdUserId}, bson.M{"$set": bson.M{"isDefault": false}})
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update addresses"})
			return
		}

		var address Address
		err = app.AddressCollection.FindOneAndUpdate(c, bson.M{"_id": addressID, "user": objectIdUserId}, bson.M{"$set": bson.M{"isDefault": true}}, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&address)
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"message": "Cannot find Address with the id: " + c.Param("id")})
			return
		} else if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update address"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Address has been set as default successfully!", "address": address})
	}
}
