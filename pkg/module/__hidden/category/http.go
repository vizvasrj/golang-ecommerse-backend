package category

import (
	"net/http"
	"src/l"
	"src/pkg/conf"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func AddCategory(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Name        string               `json:"name" binding:"required"`
			Description string               `json:"description" binding:"required"`
			Products    []primitive.ObjectID `json:"products"`
			IsActive    bool                 `json:"isActive"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "You must enter description & name."})
			return
		}

		category := Category{
			ID:          primitive.NewObjectID(),
			Name:        req.Name,
			Description: req.Description,
			Products:    req.Products,
			IsActive:    req.IsActive,
			Created:     time.Now(),
			Updated:     time.Now(),
		}

		_, err := app.CategoryCollection.InsertOne(c, category)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success":  true,
			"message":  "Category has been added successfully!",
			"category": category,
		})
	}
}

func ListCategories(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		categories, err := app.CategoryCollection.Find(c, bson.M{"isActive": true})
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}
		var categoryList []Category
		if err := categories.All(c, &categoryList); err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"categories": categoryList,
		})
	}
}

func FetchCategories(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		categories, err := app.CategoryCollection.Find(c, bson.M{})
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}
		var categoryList []Category
		if err := categories.All(c, &categoryList); err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"categories": categoryList,
		})
	}
}

func FetchCategory(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		categoryID := c.Param("id")

		category := Category{}
		err := app.CategoryCollection.FindOne(c, bson.M{"_id": categoryID}).Decode(&category)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "No Category found."})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"category": category,
		})
	}
}

func UpdateCategory(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		cID := c.Param("id")
		categoryID, err := primitive.ObjectIDFromHex(cID)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category id"})
			return
		}

		var req struct {
			Category struct {
				Slug string `json:"slug"`
			} `json:"category"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}
		update := bson.M{
			"$set": bson.M{
				"category.slug": req.Category.Slug,
			},
		}
		query := bson.M{"_id": categoryID}
		foundCategory := Category{}
		err = app.CategoryCollection.FindOne(c, bson.M{"category.slug": req.Category.Slug}).Decode(&foundCategory)
		if err != nil && err != mongo.ErrNoDocuments {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}
		if foundCategory.ID != categoryID {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Slug is already in use."})
			return
		}
		_, err = app.CategoryCollection.UpdateOne(c, query, update)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Category has been updated successfully!",
		})
	}
}

func UpdateCategoryStatus(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		categoryID := c.Param("id")
		var req struct {
			Category struct {
				IsActive bool `json:"isActive"`
			} `json:"category"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}
		query := bson.M{"_id": categoryID}
		update := bson.M{
			"$set": bson.M{
				"isActive": req.Category.IsActive,
			},
		}
		_, err := app.CategoryCollection.UpdateOne(c, query, update)
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Category has been updated successfully!",
		})
	}
}

func DeleteCategory(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		categoryID := c.Param("id")
		result, err := app.CategoryCollection.DeleteOne(c, bson.M{"_id": categoryID})
		if err != nil {
			l.DebugF("Error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}
		if result.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"message": "Category not found."})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Category has been deleted successfully!",
		})
	}
}
