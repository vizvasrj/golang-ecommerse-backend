package cart

import (
	"net/http"
	"src/l"
	"src/pkg/conf"
	"src/pkg/module/product"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func AddToCart(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var reqBody struct {
			Products []CartItem `json:"products"`
		}

		if err := c.ShouldBindJSON(&reqBody); err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		products := CalculateItemsSalesTax(reqBody.Products)

		cart := Cart{
			User:     userID.(primitive.ObjectID),
			Products: products,
		}

		cartDoc, err := app.CartCollection.InsertOne(c, cart)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}

		// err = store.DecreaseQuantity(products)
		// if err != nil {
		l.DebugF("Error: %v", err)
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
		cartID := c.Param("cartId")
		_, err := app.CartCollection.DeleteOne(c, bson.M{"_id": cartID})
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
		cartID := c.Param("cartId")
		var product product.Product

		if err := c.ShouldBindJSON(&product); err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product data"})
			return
		}

		filter := bson.M{"_id": cartID}
		update := bson.M{"$push": bson.M{"products": product}}

		_, err := app.CartCollection.UpdateOne(c, filter, update, options.Update().SetUpsert(true))
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
		cartID := c.Param("cartId")
		productID := c.Param("productId")
		product := bson.M{"product": productID}
		filter := bson.M{"_id": cartID}
		update := bson.M{"$pull": bson.M{"products": product}}
		_, err := app.CartCollection.UpdateOne(c, filter, update)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}
