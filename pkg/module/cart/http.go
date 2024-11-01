package cart

import (
	"net/http"
	"src/pkg/conf"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AddToCart(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var reqBody struct {
			Products []store.Product `json:"products"`
		}

		if err := c.ShouldBindJSON(&reqBody); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		products := store.CalculateItemsSalesTax(reqBody.Products)

		cart := store.Cart{
			User:     userID.(primitive.ObjectID),
			Products: products,
		}

		cartDoc, err := app.CartCollection.InsertOne(c, cart)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}

		err = store.DecreaseQuantity(products)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrease product quantity"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"cartId":  cartDoc.InsertedID,
		})
	}
}
