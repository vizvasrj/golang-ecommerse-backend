package order

import (
	"fmt"
	"math"
	"net/http"
	"src/common"
	"src/l"
	"src/pkg/conf"
	"src/pkg/module/cart"
	"src/pkg/module/product"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

		userID, err := primitive.ObjectIDFromHex(c.MustGet("userID").(string))
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

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

		userRoleStr := c.MustGet("role").(string)
		userRole := common.GetUserRole(userRoleStr)
		userID, err := primitive.ObjectIDFromHex(c.MustGet("userID").(string))
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

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

		for _, orderItem := range ordersDoc {
			err = app.CartCollection.FindOne(c, bson.M{"_id": orderItem.Cart}).Decode(&orderItem.Products)
			if err != nil {
				l.DebugF("Error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cart"})
				return
			}
		}

		count, err := app.OrderCollection.CountDocuments(c, bson.D{})
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"orders":      ordersDoc,
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

		userID, err := primitive.ObjectIDFromHex(c.MustGet("userID").(string))
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}
		filter := bson.M{"user": userID}
		optionsData := options.Find()
		optionsData.SetSort(bson.D{{Key: "created", Value: -1}})
		optionsData.SetLimit(int64(limitNum))
		optionsData.SetSkip(int64((pageNum - 1) * limitNum))
		cursor, err := app.OrderCollection.Find(c, filter, optionsData)
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

		for i, orderItem := range ordersDoc {
			// projection := bson.M{"products": 1, "_id": 0}
			var something cart.GetCart
			err = app.CartCollection.FindOne(
				c,
				bson.M{"_id": orderItem.Cart},
				// options.FindOne().SetProjection(projection),
			).Decode(&something)
			if err != nil {
				l.DebugF("Error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cart"})
				return
			}
			ordersDoc[i].Products = something.Products

		}

		count, err := app.OrderCollection.CountDocuments(c, filter)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"orders":      ordersDoc,
			"totalPages":  int(math.Ceil(float64(count) / float64(limitNum))),
			"currentPage": pageNum,
			"count":       count,
		})
	}
}

func FetchOrder(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		orderIDString := c.Param("orderId")
		orderID, err := primitive.ObjectIDFromHex(orderIDString)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
			return
		}

		userRoleStr := c.MustGet("role").(string)
		userRole := common.GetUserRole(userRoleStr)

		userID, err := primitive.ObjectIDFromHex(c.MustGet("userID").(string))
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}
		var orderDoc Order
		if userRole == common.RoleAdmin {
			err = app.OrderCollection.FindOne(c, bson.M{"_id": orderID}).Decode(&orderDoc)
			if err != nil {
				l.ErrorF("Error: %v", err)
				c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Cannot find order with the id: %s", orderID)})
				return
			}
		} else {
			filter := bson.M{"_id": orderID, "user": userID}
			l.InfoF("filter %#v", filter)
			err = app.OrderCollection.FindOne(c, filter).Decode(&orderDoc)
			if err != nil {
				l.ErrorF("Error: %v", err)
				l.InfoF("Order not found %s", userID)

				c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("Cannot find order with the id: %s", orderID)})
				return
			}
		}
		var cartDoc cart.GetCart
		err = app.CartCollection.FindOne(c, bson.M{"_id": orderDoc.Cart}).Decode(&cartDoc)
		if err != nil {
			l.ErrorF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cart"})
			return
		}
		// cartDoc.Products[0].Product
		var cartItem []cart.CartItem
		// var products []product.IndividualProduct
		for _, orderItem := range cartDoc.Products {
			var productDoc product.IndividualProduct
			l.InfoF("orderItem.Product %s", orderItem.Product)
			err = app.ProductCollection.FindOne(c, bson.M{"_id": orderItem.Product}).Decode(&productDoc)
			if err != nil {
				l.ErrorF("Error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product"})
				return
			}
			// products = append(products, productDoc)
			cartItem = append(cartItem, cart.CartItem{
				Product:       productDoc,
				Quantity:      orderItem.Quantity,
				PurchasePrice: orderItem.PurchasePrice,
				TotalPrice:    orderItem.TotalPrice,
				PriceWithTax:  orderItem.PriceWithTax,
				TotalTax:      orderItem.TotalTax,
				Status:        orderItem.Status,
			})
		}

		orderData := OrderGet{
			ID:       orderDoc.ID,
			Cart:     orderDoc.Cart,
			User:     orderDoc.User,
			Total:    orderDoc.Total,
			Updated:  orderDoc.Updated,
			Created:  orderDoc.Created,
			Address:  orderDoc.Address,
			Products: cartItem,
		}

		// order = store.CalculateTaxAmount(order)
		c.JSON(http.StatusOK, gin.H{"order": orderData})
	}
}

func CancelOrder(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		orderIDString := c.Param("orderId")
		userID, err := primitive.ObjectIDFromHex(c.MustGet("userID").(string))
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}
		orderID, err := primitive.ObjectIDFromHex(orderIDString)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
			return
		}

		session, err := app.MongoClient.StartSession()
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start session"})
			return
		}
		defer session.EndSession(c)

		callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
			var orderDoc Order
			filter := bson.M{"_id": orderID, "user": userID}
			l.DebugF("Order filter: %#v", filter)
			err = app.OrderCollection.FindOne(sessCtx, filter).Decode(&orderDoc)
			if err != nil {
				l.DebugF("Error finding order: %v", err)
				return nil, fmt.Errorf("cannot find order with the id: %s", orderID)
			}

			var cartDoc cart.GetCart
			err = app.CartCollection.FindOne(sessCtx, bson.M{"_id": orderDoc.Cart}).Decode(&cartDoc)
			if err != nil {
				l.DebugF("Error fetching cart: %v", err)
				return nil, fmt.Errorf("failed to fetch cart")
			}

			_, err = app.OrderCollection.DeleteOne(sessCtx, bson.M{"_id": orderID})
			if err != nil {
				l.DebugF("Error deleting order: %v", err)
				return nil, fmt.Errorf("failed to delete order")
			}

			_, err = app.CartCollection.DeleteOne(sessCtx, bson.M{"_id": orderDoc.Cart})
			if err != nil {
				l.DebugF("Error deleting cart: %v", err)
				return nil, fmt.Errorf("failed to delete cart")
			}

			return nil, nil
		}

		_, err = session.WithTransaction(c, callback, options.Transaction())
		if err != nil {
			l.DebugF("Transaction error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
