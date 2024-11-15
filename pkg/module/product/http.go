package product

import (
	"log"
	"net/http"
	"src/common"
	"src/l"
	"src/pkg/conf"
	"src/pkg/misc"
	brands "src/pkg/module/brand"
	categories "src/pkg/module/category"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetProductBySlug(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")

		var product IndividualProduct

		err := app.ProductCollection.FindOne(c, bson.M{"slug": slug, "isActive": true}).Decode(&product)
		if err != nil {
			// fmt.Println(product)
			// fmt.Println(err)
			c.JSON(http.StatusNotFound, gin.H{"message": "No product found."})
			return
		}

		c.JSON(http.StatusOK, gin.H{"product": product})
	}
}

func SearchProductsByName(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Param("name")

		filter := bson.M{
			"name":     primitive.Regex{Pattern: name, Options: "is"},
			"isActive": true,
		}
		projection := bson.M{
			"name":     1,
			"slug":     1,
			"imageUrl": 1,
			"price":    1,
			"_id":      0,
		}
		findOptions := options.Find().SetProjection(projection)

		cursor, err := app.ProductCollection.Find(c, filter, findOptions)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}
		defer cursor.Close(c)

		var products []bson.M
		if err = cursor.All(c, &products); err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}

		if len(products) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"message": "No product found."})
			return
		}

		c.JSON(http.StatusOK, gin.H{"products": products})
	}
}

func FetchStoreProductsByFilters(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		sortOrder := c.Query("sortOrder")
		rating := c.Query("rating")
		max := c.Query("max")
		min := c.Query("min")
		category := c.Query("category")
		brand := c.Query("brand")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

		var sortOrderMap bson.M
		if err := bson.UnmarshalExtJSON([]byte(sortOrder), true, &sortOrderMap); err != nil {
			log.Println(err)

			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sortOrder format"})
			return
		}

		categoryFilter := bson.M{}
		if category != "" {
			categoryFilter["category"] = category
		}

		basicQuery := getStoreProductsQuery(min, max, rating)

		categoryDoc := categories.Category{}
		if err := app.CategoryCollection.FindOne(c, bson.M{"slug": categoryFilter["category"], "isActive": true}).Decode(&categoryDoc); err == nil {
			basicQuery = append(basicQuery, bson.M{
				"$match": bson.M{
					"isActive": true,
					"_id": bson.M{
						"$in": categoryDoc.Products,
					},
				},
			})
		}

		brandDoc := brands.Brand{}
		if err := app.BrandCollection.FindOne(c, bson.M{"slug": brand, "isActive": true}).Decode(&brandDoc); err == nil {
			basicQuery = append(basicQuery, bson.M{
				"$match": bson.M{
					"brand._id": brandDoc.ID,
				},
			})
		}

		productsCount, err := app.ProductCollection.Aggregate(c, basicQuery)
		if err != nil {
			log.Println(err)
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}

		var count int
		for productsCount.Next(c) {
			count++
		}

		size := 0
		if count > limit {
			size = page - 1
		}
		currentPage := 1
		if count > limit {
			currentPage = page
		}

		paginateQuery := []bson.M{
			{"$sort": sortOrderMap},
			{"$skip": size * limit},
			{"$limit": limit},
		}

		// var products []bson.M
		// if userDoc != nil {
		// 	wishListQuery := getStoreProductsWishListQuery(userDoc.ID)
		// 	products, err = app.ProductCollection.Aggregate(c, append(wishListQuery, append(basicQuery, paginateQuery...)...))
		// } else {
		// 	products, err = app.ProductCollection.Aggregate(c, append(basicQuery, paginateQuery...))
		// }

		productsCursor, err := app.ProductCollection.Aggregate(c, append(basicQuery, paginateQuery...))
		if err != nil {
			log.Println(err)

			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}

		// var demo []bson.M
		// if err = productsCursor.All(c, &demo); err != nil {
		// 	log.Println(err)

		l.DebugF("Error: %v", err)
		// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Your request could not be processed. Please try again."})
		// 	return
		// }
		// fmt.Printf("%#v\n", demo)

		var products []Product
		if err = productsCursor.All(c, &products); err != nil {
			log.Println(err)

			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"products":    products,
			"totalPages":  (count + limit - 1) / limit,
			"currentPage": currentPage,
			"count":       count,
		})
	}
}

func getStoreProductsQuery(min, max, rating string) []bson.M {
	minPrice, _ := strconv.ParseFloat(min, 64)
	maxPrice, _ := strconv.ParseFloat(max, 64)
	ratingValue, _ := strconv.ParseFloat(rating, 64)

	priceFilter := bson.M{}
	if minPrice > 0 && maxPrice > 0 {
		priceFilter = bson.M{"price": bson.M{"$gte": minPrice, "$lte": maxPrice}}
	}

	ratingFilter := bson.M{"rating": bson.M{"$gte": ratingValue}}

	matchQuery := bson.M{
		"isActive":      true,
		"price":         priceFilter["price"],
		"averageRating": ratingFilter["rating"],
	}

	basicQuery := []bson.M{
		{
			"$lookup": bson.M{
				"from":         "brands",
				"localField":   "brand",
				"foreignField": "_id",
				"as":           "brands",
			},
		},
		{
			"$unwind": bson.M{
				"path":                       "$brands",
				"preserveNullAndEmptyArrays": true,
			},
		},
		{
			"$addFields": bson.M{
				"brand.name":     "$brands.name",
				"brand._id":      "$brands._id",
				"brand.isActive": "$brands.isActive",
			},
		},
		{
			"$match": bson.M{
				"brand.isActive": true,
			},
		},
		{
			"$lookup": bson.M{
				"from":         "reviews",
				"localField":   "_id",
				"foreignField": "product",
				"as":           "reviews",
			},
		},
		{
			"$addFields": bson.M{
				"totalRatings": bson.M{"$sum": "$reviews.rating"},
				"totalReviews": bson.M{"$size": "$reviews"},
			},
		},
		{
			"$addFields": bson.M{
				"averageRating": bson.M{
					"$cond": bson.A{
						bson.M{"$eq": bson.A{"$totalReviews", 0}},
						0,
						bson.M{"$divide": bson.A{"$totalRatings", "$totalReviews"}},
					},
				},
			},
		},
		{
			"$match": matchQuery,
		},
		{
			"$project": bson.M{
				"brands":  0,
				"reviews": 0,
			},
		},
	}

	return basicQuery
}

func FetchProductNames(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var products []Product
		projection := bson.M{"name": 1}

		cursor, err := app.ProductCollection.Find(c, bson.M{}, options.Find().SetProjection(projection))
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}
		defer cursor.Close(c)

		if err = cursor.All(c, &products); err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}

		c.JSON(http.StatusOK, gin.H{"products": products})
	}
}

func AddProduct(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input AddProductInput

		if err := c.ShouldBind(&input); err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		file, err := c.FormFile("image")
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Image upload failed"})
			return
		}

		foundProduct := app.ProductCollection.FindOne(c, bson.M{"sku": input.SKU})
		if foundProduct.Err() != mongo.ErrNoDocuments {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "This SKU is already in use."})
			return
		}

		// ! use diffrent goroutine for this.
		imageUrl, imageKey, err := misc.S3Upload(file, app)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Image upload failed"})
			return
		}

		product := bson.M{
			"sku":         input.SKU,
			"name":        input.Name,
			"description": input.Description,
			"quantity":    input.Quantity,
			"price":       input.Price,
			"taxable":     input.Taxable,
			"isActive":    input.IsActive,
			"brand":       input.Brand,
			"imageUrl":    imageUrl,
			"imageKey":    imageKey,
			"merchant":    c.MustGet("user").(primitive.ObjectID),
		}

		_, err = app.ProductCollection.InsertOne(c, product)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Product has been added successfully!",
			"product": product,
		})
	}
}

func FetchProducts(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")

		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		userRole := common.GetUserRole(role)

		var products []IndividualProduct

		if userRole == common.RoleMerchant {
			mID := c.MustGet("merchantID").(string)
			l.DebugF("merchant ID: %#v", mID)
			merchantID, err := primitive.ObjectIDFromHex(mID)
			if err != nil {
				l.DebugF("Error: %v", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid merchant ID"})
				return
			}

			filter := bson.M{"merchant": merchantID}
			cursor, err := app.ProductCollection.Find(c, filter)
			if err != nil {
				l.DebugF("Error: %v", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
				return
			}
			if err = cursor.All(c, &products); err != nil {
				l.DebugF("Error: %v", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
				return
			}
		} else {
			cursor, err := app.ProductCollection.Find(c, bson.M{})
			if err != nil {
				l.DebugF("Error: %v", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
				return
			}
			if err = cursor.All(c, &products); err != nil {
				l.DebugF("Error: %v", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"products": products})
	}
}

func FetchProduct(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		_userID, exists := c.MustGet("merchantID").(string)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		merchantID, err := primitive.ObjectIDFromHex(_userID)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid merchant ID"})
			return
		}

		userRole, ok := c.MustGet("role").(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "failed to get user role"})
			return
		}

		role := common.GetUserRole(userRole)

		productId := c.Param("id")
		objectId, err := primitive.ObjectIDFromHex(productId)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
			return
		}

		var productDoc IndividualProduct

		if role == common.RoleMerchant {
			filter := bson.M{"_id": objectId, "merchant": merchantID}
			err = app.ProductCollection.FindOne(c, filter).Decode(&productDoc)
		} else {
			filter := bson.M{"_id": objectId}
			err = app.ProductCollection.FindOne(c, filter).Decode(&productDoc)
		}

		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{"message": "No product found."})
			} else {
				l.DebugF("Error: %v", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"product": productDoc})
	}
}

func UpdateProduct(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.MustGet("userID").(primitive.ObjectID)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		userRole := c.MustGet("role").(common.UserRole)

		productId := c.Param("id")
		objectId, err := primitive.ObjectIDFromHex(productId)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
			return
		}

		var update AddProductInput
		if err := c.ShouldBindJSON(&update); err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		sku := update.SKU

		var foundProduct Product
		filter := bson.M{"sku": sku}
		err = app.ProductCollection.FindOne(c, filter).Decode(&foundProduct)
		if err == nil && foundProduct.ID != objectId {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Sku or slug is already in use."})
			return
		}

		// Check if the user is a merchant and if the product belongs to the merchant
		if userRole == common.RoleMerchant {
			var product Product
			err = app.ProductCollection.FindOne(c, bson.M{"_id": objectId, "merchant": userID}).Decode(&product)
			if err == mongo.ErrNoDocuments {
				l.DebugF("Error: %v", err)
				c.JSON(http.StatusForbidden, gin.H{"error": "You can only update products that belong to your merchant account."})
				return
			} else if err != nil {
				l.DebugF("Error: %v", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
				return
			}
		}

		query := bson.M{"_id": objectId}
		updateResult := app.ProductCollection.FindOneAndUpdate(c, query, bson.M{"$set": update}, nil)
		if updateResult.Err() != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Product has been updated successfully!",
		})
	}
}

func UpdateProductStatus(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.MustGet("userID").(primitive.ObjectID)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		userRole, exists := c.MustGet("role").(common.UserRole)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "failed to get user role"})
			return
		}

		productId := c.Param("id")
		objectId, err := primitive.ObjectIDFromHex(productId)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
			return
		}

		var update bson.M
		if err := c.ShouldBindJSON(&update); err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		// Check if the user is a merchant and if the product belongs to the merchant
		if userRole == common.RoleMember {
			var product Product
			err = app.ProductCollection.FindOne(c, bson.M{"_id": objectId, "merchant": userID}).Decode(&product)
			if err == mongo.ErrNoDocuments {
				l.DebugF("Error: %v", err)
				c.JSON(http.StatusForbidden, gin.H{"error": "You can only update products that belong to your merchant account."})
				return
			} else if err != nil {
				l.DebugF("Error: %v", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
				return
			}
		}

		query := bson.M{"_id": objectId}
		updateResult := app.ProductCollection.FindOneAndUpdate(c, query, bson.M{"$set": update}, nil)
		if updateResult.Err() != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Product has been updated successfully!",
		})
	}
}

func DeleteProduct(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.MustGet("userID").(primitive.ObjectID)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		userRole, exists := c.MustGet("role").(common.UserRole)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to get user role"})
			return
		}

		productId := c.Param("id")
		objectId, err := primitive.ObjectIDFromHex(productId)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
			return
		}

		// Check if the user is a merchant and if the product belongs to the merchant
		if userRole == common.RoleMerchant {
			var product Product
			err = app.ProductCollection.FindOne(c, bson.M{"_id": objectId, "merchant": userID}).Decode(&product)
			if err == mongo.ErrNoDocuments {
				l.DebugF("Error: %v", err)
				c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
				return
			} else if err != nil {
				l.DebugF("Error: %v", err)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
				return
			}
		}

		// Delete the product
		deleteResult, err := app.ProductCollection.DeleteOne(c, bson.M{"_id": objectId})
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}

		if deleteResult.DeletedCount == 0 {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Product has been deleted successfully!",
		})
	}
}
