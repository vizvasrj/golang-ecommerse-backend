package cart

import (
	"net/http"
	"src/l"
	"src/pkg/conf"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func AddToCart(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var cartProduct AddProductToCartRequest
		if err := c.ShouldBindJSON(&cartProduct); err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		// l.DebugF("a %v", a)

		userID, err := primitive.ObjectIDFromHex(c.GetString("userID"))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}
		// l.DebugF("reqBody %v", reqBody)
		// products := CalculateItemsSalesTax(reqBody.Products)

		cart := GetCart{
			ID:       primitive.NewObjectID(),
			User:     userID,
			Products: []GetCartItem{{Product: cartProduct.Product, Quantity: cartProduct.Quantity}},
			Updated:  time.Now(),
			Created:  time.Now(),
		}
		l.DebugF("cart %v", cart)

		cartDoc, err := app.CartCollection.InsertOne(c, cart)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}

		// err = store.DecreaseQuantity(products)
		// if err != nil {
		// l.DebugF("Error: %v", err)
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrease product quantity"})
		// 	return
		// }

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"cartId":  cartDoc.InsertedID,
		})
	}
}

func DeleteCart(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		cartIDhex := c.Param("cartId")
		cartID, err := primitive.ObjectIDFromHex(cartIDhex)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cart ID"})
			return
		}
		userID, err := primitive.ObjectIDFromHex(c.GetString("userID"))
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		count, err := app.CartCollection.CountDocuments(c, bson.M{"_id": cartID, "user": userID})
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		if count == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
			return
		}

		_, err = app.CartCollection.DeleteOne(c, bson.M{"_id": cartID})
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}

func AddProductToCart(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		cartIDhex := c.Param("cartId")
		cartID, err := primitive.ObjectIDFromHex(cartIDhex)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cart ID"})
			return
		}

		var cart GetCart
		cartFilter := bson.M{"_id": cartID}
		err = app.CartCollection.FindOne(c, cartFilter).Decode(&cart)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cart not found"})
			return
		}
		userID, err := primitive.ObjectIDFromHex(c.GetString("userID"))
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}
		if cart.User != userID {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
			return
		}

		var cartItem GetCartItem

		if err := c.ShouldBindJSON(&cartItem); err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product data"})
			return
		}

		// verify product exists or not
		count, err := app.ProductCollection.CountDocuments(c, bson.M{"_id": cartItem.Product})
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
			return
		}
		if count == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product not found"})
			return
		}

		filter := bson.M{"_id": cartID}
		update := bson.M{"$push": bson.M{"products": cartItem}}

		_, err = app.CartCollection.UpdateOne(c, filter, update, options.Update().SetUpsert(true))
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}

func RemoveProductFromCart(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		cartIDhex := c.Param("cartId")
		cartID, err := primitive.ObjectIDFromHex(cartIDhex)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cart ID"})
			return
		}

		var cart GetCart
		cartFilter := bson.M{"_id": cartID}
		err = app.CartCollection.FindOne(c, cartFilter).Decode(&cart)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cart not found"})
			return
		}
		userID, err := primitive.ObjectIDFromHex(c.GetString("userID"))
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}
		if cart.User != userID {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
			return
		}

		productIDhex := c.Param("productId")
		productID, err := primitive.ObjectIDFromHex(productIDhex)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
			return
		}
		product := bson.M{"product": productID}
		filter := bson.M{"_id": cartID}
		update := bson.M{"$pull": bson.M{"products": product}}

		_, err = app.CartCollection.UpdateOne(c, filter, update)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}
