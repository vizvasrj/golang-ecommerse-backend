package order

import (
	"fmt"
	"math"
	"net/http"
	"src/common"
	"src/l"
	"src/pkg/conf"
	"src/pkg/module/cart"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func AddOrder(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			CartID  primitive.ObjectID `json:"cartId"`
			Total   float64            `json:"total"`
			Address Address            `json:"address"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		userID := c.MustGet("userID").(primitive.ObjectID)

		order := Order{
			Cart:    req.CartID,
			User:    userID,
			Total:   req.Total,
			Address: req.Address,
		}

		orderDoc, err := app.OrderCollection.InsertOne(c, order)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to place order"})
			return
		}
		var cartDoc cart.Cart
		err = app.CartCollection.FindOne(c, bson.M{"_id": order.Cart}).Decode(&cartDoc)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cart"})
			return
		}

		// newOrder := gin.H{
		// 	"_id":      orderDoc.InsertedID,
		// 	"created":  order.Created,
		// 	"user":     order.User,
		// 	"total":    order.Total,
		// 	"products": cartDoc.Products,
		// }

		// Assuming mailgun.SendEmail is a function that sends an email
		// err = mailgun.SendEmail(order.UserID, "order-confirmation", newOrder)
		// if err != nil {
		l.DebugF("Error: %v", err)
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send confirmation email"})
		// 	return
		// }

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Your order has been placed successfully!",
			"order":   gin.H{"_id": orderDoc.InsertedID},
		})
	}
}

func SearchOrders(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		search := c.Query("search")

		objectID, err := primitive.ObjectIDFromHex(search)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid search query"})
			return
		}

		var ordersDoc []Order

		userRole := c.MustGet("role").(common.UserRole)
		userID := c.MustGet("userID").(primitive.ObjectID)

		filter := bson.M{"_id": objectID}
		if userRole != common.RoleAdmin {
			filter["user"] = userID
		}

		cursor, err := app.OrderCollection.Find(c, filter)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}
		defer cursor.Close(c)

		for cursor.Next(c) {
			var order Order
			if err := cursor.Decode(&order); err != nil {
				l.DebugF("Error: %v", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
				return
			}
			ordersDoc = append(ordersDoc, order)
		}

		c.JSON(http.StatusOK, gin.H{"orders": ordersDoc})
	}
}

func FetchOrders(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		page := c.DefaultQuery("page", "1")
		limit := c.DefaultQuery("limit", "10")

		pageNum, err := strconv.Atoi(page)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
			return
		}

		limitNum, err := strconv.Atoi(limit)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit"})
			return
		}

		options := options.Find()
		options.SetSort(bson.D{{Key: "created", Value: -1}})
		options.SetLimit(int64(limitNum))
		options.SetSkip(int64((pageNum - 1) * limitNum))

		cursor, err := app.OrderCollection.Find(c, bson.D{}, options)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}
		defer cursor.Close(c)

		var ordersDoc []Order
		if err := cursor.All(c, &ordersDoc); err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}

		count, err := app.OrderCollection.CountDocuments(c, bson.D{})
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}

		orders := ordersDoc

		c.JSON(http.StatusOK, gin.H{
			"orders":      orders,
			"totalPages":  int(math.Ceil(float64(count) / float64(limitNum))),
			"currentPage": pageNum,
			"count":       count,
		})
	}
}

func FetchUserOrders(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		page := c.DefaultQuery("page", "1")
		limit := c.DefaultQuery("limit", "10")
		pageNum, err := strconv.Atoi(page)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page number"})
			return
		}

		limitNum, err := strconv.Atoi(limit)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit"})
			return
		}

		userID := c.MustGet("userID").(primitive.ObjectID)
		filter := bson.M{"user": userID}
		options := options.Find()
		options.SetSort(bson.D{{Key: "created", Value: -1}})
		options.SetLimit(int64(limitNum))
		options.SetSkip(int64((pageNum - 1) * limitNum))
		cursor, err := app.OrderCollection.Find(c, filter, options)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}
		defer cursor.Close(c)
		var ordersDoc []Order
		if err := cursor.All(c, &ordersDoc); err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}
		count, err := app.OrderCollection.CountDocuments(c, filter)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}
		orders := ordersDoc
		c.JSON(http.StatusOK, gin.H{
			"orders":      orders,
			"totalPages":  int(math.Ceil(float64(count) / float64(limitNum))),
			"currentPage": pageNum,
			"count":       count,
		})
	}
}

func FetchOrder(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("orderId")
		userRole := c.MustGet("role").(common.UserRole)
		userID := c.MustGet("userID").(primitive.ObjectID)
		var orderDoc Order
		var err error
		if userRole == common.RoleAdmin {
			err = app.OrderCollection.FindOne(c, bson.M{"_id": orderID}).Decode(&orderDoc)
		} else {
			err = app.OrderCollection.FindOne(c, bson.M{"_id": orderID, "user": userID}).Decode(&orderDoc)
		}
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Cannot find order with the id: %s", orderID)})
			return
		}
		var cartDoc cart.Cart
		err = app.CartCollection.FindOne(c, bson.M{"_id": orderDoc.Cart}).Decode(&cartDoc)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cart"})
			return
		}
		order := Order{
			ID:      orderDoc.ID,
			Total:   orderDoc.Total,
			Created: orderDoc.Created,
			// TotalTax:   0,
			// Products: cartDoc.Products,
			Cart:    orderDoc.Cart,
			Address: orderDoc.Address,
		}
		// order = store.CalculateTaxAmount(order)
		c.JSON(http.StatusOK, gin.H{"order": order})
	}
}

func CancelOrder(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("orderId")
		userID := c.MustGet("userID").(primitive.ObjectID)
		var orderDoc Order
		err := app.OrderCollection.FindOne(c, bson.M{"_id": orderID, "user": userID}).Decode(&orderDoc)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Cannot find order with the id: %s", orderID)})
			return
		}
		var cartDoc cart.Cart
		err = app.CartCollection.FindOne(c, bson.M{"_id": orderDoc.Cart}).Decode(&cartDoc)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cart"})
			return
		}
		// increaseQuantity(cartDoc.Products)

		_, err = app.OrderCollection.DeleteOne(c, bson.M{"_id": orderID})
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete order"})
			return
		}
		_, err = app.CartCollection.DeleteOne(c, bson.M{"_id": orderDoc.Cart})
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete cart"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
		})
	}
}

func UpdateItemStatus(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {

		_item_id := c.Param("itemId")
		itemId, err := primitive.ObjectIDFromHex(_item_id)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
			return
		}
		requestData := struct {
			OrderId primitive.ObjectID  `json:"orderId"`
			CartId  primitive.ObjectID  `json:"cartId"`
			Status  cart.CartItemStatus `json:"status"`
		}{}
		var status cart.CartItemStatus

		orderId := requestData.OrderId
		cartId := requestData.CartId
		if requestData.Status == "" {
			status = cart.Cancelled
		}

		var foundCart cart.Cart
		err = app.CartCollection.FindOne(c, bson.M{"products._id": itemId}).Decode(&foundCart)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cart not found"})
			return
		}

		var foundCartProduct cart.CartItem
		for _, product := range foundCart.Products {
			if product.Product.ID == itemId {
				foundCartProduct = product
				break
			}
		}

		_, err = app.CartCollection.UpdateOne(
			c,
			bson.M{"products._id": itemId},
			bson.M{"$set": bson.M{"products.$.status": status}},
		)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update item status"})
			return
		}

		if status == cart.Cancelled {
			_, err = app.ProductCollection.UpdateOne(
				c,
				bson.M{"_id": foundCartProduct.Product},
				bson.M{"$inc": bson.M{"quantity": foundCartProduct.Quantity}},
			)
			if err != nil {
				l.DebugF("Error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product quantity"})
				return
			}

			var cartData cart.Cart
			err = app.CartCollection.FindOne(c, bson.M{"_id": cartId}).Decode(&cartData)
			if err != nil {
				l.DebugF("Error: %v", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Cart not found"})
				return
			}

			cancelledItems := 0
			for _, item := range cartData.Products {
				if item.Status == cart.Cancelled {
					cancelledItems++
				}
			}

			if len(cartData.Products) == cancelledItems {
				_, err = app.OrderCollection.DeleteOne(c, bson.M{"_id": orderId})
				if err != nil {
					l.DebugF("Error: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel order"})
					return
				}
				_, err = app.CartCollection.DeleteOne(c, bson.M{"_id": cartId})
				if err != nil {
					l.DebugF("Error: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete cart"})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"success":        true,
					"orderCancelled": true,
					"message":        "Order has been cancelled successfully",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Item has been cancelled successfully!",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Item status has been updated successfully!",
		})
	}
}
