package review

import (
	"math"
	"net/http"
	"src/common"
	"src/l"
	"src/pkg/conf"
	"src/pkg/module/product"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func AddReview(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var review PutReviewInput
		if err := c.ShouldBindJSON(&review); err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}
		l.InfoF("Review: %#v", review)
		rating, err := strconv.ParseFloat(review.Rating, 64)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rating"})
			return
		}
		var newReview Review
		newReview.Title = review.Title
		newReview.Rating = rating
		newReview.Review = review.Review
		newReview.IsRecommended = review.IsRecommended
		newReview.Product = review.Product

		userID, ok := c.MustGet("userID").(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		objectId, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			l.DebugF("Invalid user ID: %s", userID)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		reviewUser := ReviewUser{}
		firstName := c.GetString("firstname")
		lastName := c.GetString("lastname")
		email := c.GetString("email")

		reviewUser.ID = objectId
		reviewUser.FirstName = firstName
		reviewUser.LastName = lastName
		reviewUser.Email = email

		newReview.User = reviewUser
		newReview.Created = time.Now()
		newReview.Updated = time.Now()
		newReview.Status = WaitingApproval

		_, err = app.ReviewCollection.InsertOne(c, newReview)
		if err != nil {
			l.ErrorF("Could not add review: %s", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not add review"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Your review has been added successfully and will appear when approved!",
			"review":  newReview,
		})
	}
}

func GetAllReviews(app *conf.Config) gin.HandlerFunc {
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

		skip := (page - 1) * limit
		findOptions := options.Find()
		findOptions.SetSort(bson.D{{Key: "created", Value: -1}})
		findOptions.SetLimit(int64(limit))
		findOptions.SetSkip(int64(skip))

		cursor, err := app.ReviewCollection.Find(c, bson.M{}, findOptions)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}
		defer cursor.Close(c)

		var reviews []Review
		if err := cursor.All(c, &reviews); err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding reviews"})
			return
		}

		count, err := app.ReviewCollection.CountDocuments(c, bson.M{})
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error counting reviews"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"reviews":     reviews,
			"totalPages":  math.Ceil(float64(count) / float64(limit)),
			"currentPage": page,
			"count":       count,
		})
	}
}

func GetProductReviewsBySlug(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")

		var productDoc product.IndividualProduct
		err := app.ProductCollection.FindOne(c, bson.M{"slug": slug}).Decode(&productDoc)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusNotFound, gin.H{"message": "No product found."})
			return
		}

		filter := bson.M{
			"product": productDoc.ID,
			"status":  Approved,
		}
		findOptions := options.Find().SetSort(bson.D{{Key: "created", Value: -1}})
		cursor, err := app.ReviewCollection.Find(c, filter, findOptions)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}
		defer cursor.Close(c)
		// something := []bson.M{}
		// if err := cursor.All(c, &something); err != nil {
		// 	l.DebugF("Error: %v", err)
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding reviews"})
		// 	return
		// }
		// l.DebugF("Reviews: %s", something)
		reviews := []Review{}
		if err := cursor.All(c, &reviews); err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding reviews"})
			return
		}

		// l.DebugF("Reviews: %#v", reviews)

		c.JSON(http.StatusOK, gin.H{"reviews": reviews})
	}
}

func UpdateReview(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := primitive.ObjectIDFromHex(c.MustGet("userID").(string))
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		reviewId := c.Param("id")
		objectId, err := primitive.ObjectIDFromHex(reviewId)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
			return
		}

		var update PutReviewInput
		if err := c.ShouldBindJSON(&update); err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}
		rating, err := strconv.ParseFloat(update.Rating, 64)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rating"})
			return
		}
		if rating < 1 || rating > 5 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Rating must be between 1 and 5"})
			return
		}

		// Check if the user is the owner of the review
		review := Review{}
		err = app.ReviewCollection.FindOne(c, bson.M{"_id": objectId}).Decode(&review)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
			return
		}

		if review.User.ID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to update this review"})
			return
		}

		update.Updated = time.Now()
		var updateReview = bson.M{
			"title":         update.Title,
			"rating":        rating,
			"review":        update.Review,
			"isRecommended": update.IsRecommended,
			"updated":       update.Updated,
		}

		query := bson.M{"_id": objectId}
		updateResult := app.ReviewCollection.FindOneAndUpdate(c, query, bson.M{"$set": updateReview}, options.FindOneAndUpdate().SetReturnDocument(options.After))
		if updateResult.Err() != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Review has been updated successfully!",
		})
	}
}

func ApproveReview(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		reviewId := c.Param("reviewId")
		objectId, err := primitive.ObjectIDFromHex(reviewId)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
			return
		}

		merchantID, ok := c.MustGet("merchantID").(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Merchant not authenticated"})
			return
		}

		merchantObjectId, err := primitive.ObjectIDFromHex(merchantID)
		if err != nil {
			l.DebugF("Invalid merchant ID: %s", merchantID)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid merchant ID"})
			return
		}

		review := Review{}
		err = app.ReviewCollection.FindOne(c, bson.M{"_id": objectId}).Decode(&review)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Review not found"})
			return
		}

		l.DebugF("review: product id: %s", review.Product)

		findProduct := product.IndividualProduct{}
		err = app.ProductCollection.FindOne(c, bson.M{"_id": review.Product}).Decode(&findProduct)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product not found"})
			return
		}
		role := c.MustGet("role")
		roleType := common.GetUserRole(role)

		if findProduct.Merchant != merchantObjectId && roleType != common.RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to approve this review"})
			return
		}

		query := bson.M{"_id": objectId}
		update := bson.M{
			"$set": bson.M{
				"status":   Approved,
				"isActive": true,
			},
		}

		updateResult := app.ReviewCollection.FindOneAndUpdate(c, query, update, options.FindOneAndUpdate().SetReturnDocument(options.After))
		if updateResult.Err() != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
		})
	}
}

func RejectReview(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		reviewId := c.Param("reviewId")
		objectId, err := primitive.ObjectIDFromHex(reviewId)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
			return
		}

		merchantID, ok := c.MustGet("merchantID").(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Merchant not authenticated"})
			return
		}

		merchantObjectId, err := primitive.ObjectIDFromHex(merchantID)
		if err != nil {
			l.DebugF("Invalid merchant ID: %s", merchantID)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid merchant ID"})
			return
		}

		review := Review{}
		err = app.ReviewCollection.FindOne(c, bson.M{"_id": objectId}).Decode(&review)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Review not found"})
			return
		}

		findProduct := product.Product{}
		err = app.ProductCollection.FindOne(c, bson.M{"_id": review.Product}).Decode(&findProduct)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product not found"})
			return
		}

		role := c.MustGet("role")
		roleType := common.GetUserRole(role)

		if findProduct.Merchant != merchantObjectId && roleType != common.RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to reject this review"})
			return
		}

		query := bson.M{"_id": objectId}
		update := bson.M{
			"$set": bson.M{
				"status":   Rejected,
				"isActive": false,
			},
		}

		updateResult := app.ReviewCollection.FindOneAndUpdate(c, query, update, options.FindOneAndUpdate().SetReturnDocument(options.After))
		if updateResult.Err() != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
		})
	}
}

func DeleteReview(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		reviewID := c.Param("id")
		objectID, err := primitive.ObjectIDFromHex(reviewID)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review id"})
			return
		}

		// Retrieve the review from the database
		review := Review{}
		err = app.ReviewCollection.FindOne(c, bson.M{"_id": objectID}).Decode(&review)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Review not found"})
			return
		}

		// Retrieve the user ID from the context
		userID, ok := c.MustGet("userID").(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			return
		}

		userIDObjectID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			l.DebugF("Invalid user ID: %s", userID)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		role := c.MustGet("role")
		roleType := common.GetUserRole(role)

		deleteReview := func() {
			result, err := app.ReviewCollection.DeleteOne(c, bson.M{"_id": objectID})
			if err != nil {
				l.DebugF("Error: %v", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
				return
			}
			l.InfoF("Review deleted: %v", result)
		}

		if userIDObjectID == review.User.ID {
			// Proceed to delete the review
			deleteReview()

		} else if roleType == common.RoleMerchant {
			// Check if the review belongs to the merchant
			product := product.Product{}
			err = app.ProductCollection.FindOne(c, bson.M{"_id": review.Product}).Decode(&product)
			if err != nil {
				l.DebugF("Error: %v", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Product not found"})
				return
			}
			merchantID, err := primitive.ObjectIDFromHex(c.MustGet("merchantID").(string))
			if err != nil {
				l.DebugF("Invalid merchant ID: %s", c.MustGet("merchantID").(string))
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid merchant ID"})
				return
			}

			if product.Merchant != merchantID {
				c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to delete this review"})
				return
			}

			// Proceed to delete the review
			deleteReview()

		} else if roleType == common.RoleAdmin {
			// Proceed to delete the review
			deleteReview()
		} else {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to delete this review"})
			return
		}

		// Check if the review belongs to the user or if the user is an admin
		if review.User.ID != userIDObjectID && roleType != common.RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to delete this review"})
			return
		}

		// Proceed to delete the review
		result, err := app.ReviewCollection.DeleteOne(c, bson.M{"_id": objectID})
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Review has been deleted successfully!",
			"review":  result,
		})
	}
}
