package cart

import (
	"fmt"
	"net/http"
	"src/l"
	"src/pkg/conf"
	"src/pkg/module/product"
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

// this function will do every thing
func AddProductToCartV2(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract cart ID and user ID
		cartID, err := extractCartID(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cart ID"})
			return
		}
		l.DebugF("cartID %v", cartID)

		userID, err := extractUserID(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}
		l.DebugF("userID %v", userID)

		// Parse the cart item from request
		cartItem, err := parseCartItem(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		l.DebugF("cartItem %v", cartItem)

		// Verify product existence
		if err := verifyProductExists(app, c, cartItem.Product); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		l.DebugF("product exists")

		if cartID.IsZero() {
			// Create a new cart if no cart ID is provided
			handleNewCart(app, c, userID, cartItem)
			return
		}
		l.DebugF("cartID not zero")

		// Handle existing cart logic
		handleExistingCart(app, c, cartID, userID, cartItem)
	}
}

func extractCartID(c *gin.Context) (primitive.ObjectID, error) {
	cartIDhex := c.Query("cartId")
	if cartIDhex == "" {
		return primitive.NilObjectID, nil // No cart ID provided
	}
	return primitive.ObjectIDFromHex(cartIDhex)
}

func extractUserID(c *gin.Context) (primitive.ObjectID, error) {
	userIDStr := c.GetString("userID")
	if userIDStr == "" {
		return primitive.NilObjectID, nil // User not logged in
	}
	return primitive.ObjectIDFromHex(userIDStr)
}

func parseCartItem(c *gin.Context) (GetCartItem, error) {
	var cartItem GetCartItem
	if err := c.ShouldBindJSON(&cartItem); err != nil {
		return cartItem, fmt.Errorf("invalid product data")
	}
	return cartItem, nil
}

func verifyProductExists(app *conf.Config, c *gin.Context, productID primitive.ObjectID) error {
	count, err := app.ProductCollection.CountDocuments(c, bson.M{"_id": productID})
	if err != nil || count == 0 {
		return fmt.Errorf("product not found")
	}
	return nil
}

func handleNewCart(app *conf.Config, c *gin.Context, userID primitive.ObjectID, cartItem GetCartItem) {
	newCart := GetCart{
		ID:       primitive.NewObjectID(),
		User:     userID,
		Products: []GetCartItem{cartItem},
	}

	_, err := app.CartCollection.InsertOne(c, newCart)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Cart created and product added", "cartId": newCart.ID.Hex()})
}

func handleExistingCart(app *conf.Config, c *gin.Context, cartID, userID primitive.ObjectID, cartItem GetCartItem) {
	var cart GetCart
	err := app.CartCollection.FindOne(c, bson.M{"_id": cartID}).Decode(&cart)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cart not found"})
		return
	}

	if !cart.User.IsZero() && cart.User != userID {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
		return
	}

	if cart.User.IsZero() && !userID.IsZero() {
		// Associate the cart with the logged-in user
		_, err = app.CartCollection.UpdateOne(
			c,
			bson.M{"_id": cartID},
			bson.M{"$set": bson.M{"user": userID}},
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to associate cart with user"})
			return
		}
	}

	// Check if the product already exists in the cart
	found := false
	for i, existingItem := range cart.Products {
		if existingItem.Product == cartItem.Product {
			// Product already exists, update its quantity
			cart.Products[i].Quantity += cartItem.Quantity
			found = true
			break
		}
	}

	if found {
		// Update the cart with the modified product list
		_, err = app.CartCollection.UpdateOne(
			c,
			bson.M{"_id": cartID},
			bson.M{
				"$set": bson.M{
					"products": cart.Products,
					"updated":  time.Now(),
				},
			},
		)
	} else {
		// Product does not exist, append it to the cart
		_, err = app.CartCollection.UpdateOne(
			c,
			bson.M{"_id": cartID},
			bson.M{
				"$push": bson.M{"products": cartItem},
				"$set":  bson.M{"updated": time.Now()},
			},
		)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cart"})
		return
	}

	err = app.CartCollection.FindOne(c, bson.M{"_id": cartID}).Decode(&cart)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve updated cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "cart": cart, "cartId": cart.ID.Hex()})
}

func GetCartByCartID(app *conf.Config) gin.HandlerFunc {
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
		userIDStr := c.GetString("userID")
		if userIDStr != "" {
			userID, err := primitive.ObjectIDFromHex(userIDStr)
			if err != nil {
				l.DebugF("Error: %v", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
				return
			}

			if !cart.User.IsZero() && cart.User != userID {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized access"})
				return
			}
		}

		// get products details that if product is deleted from product collection or not
		// also need price of that product
		total := 0.0
		var products []product.IndividualProduct
		for _, _product := range cart.Products {
			var productDoc product.IndividualProduct
			err = app.ProductCollection.FindOne(c, bson.M{"_id": _product.Product}).Decode(&productDoc)
			if err != nil {
				l.DebugF("Error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
				return
			}
			totalPrice := productDoc.Price * float64(_product.Quantity)
			total += totalPrice
			productDoc.TotalPrice = totalPrice
			productDoc.Quantity = _product.Quantity
			products = append(products, productDoc)

		}

		c.JSON(http.StatusOK, gin.H{"success": true, "cart": products, "total": total})
	}
}
