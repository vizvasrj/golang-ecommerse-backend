package brand

import (
	"net/http"
	"src/pkg/conf"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AddBrand(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var brand Brand
		if err := c.ShouldBindJSON(&brand); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		if brand.Name == "" || brand.Description == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You must enter description & name."})
			return
		}

		_, err := app.BrandCollection.InsertOne(c, brand)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Brand has been added successfully!", "brand": brand})
	}
}

// fetch store brands api
func ListBrands(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		brands, err := app.BrandCollection.Find(c, bson.M{"isActive": true})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}
		c.JSON(http.StatusOK, gin.H{"brands": brands})
	}
}

func GetBrands(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		brands, err := app.BrandCollection.Find(c, bson.M{"isActive": true})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}
		c.JSON(http.StatusOK, gin.H{"brands": brands})
	}
}

// get brand by ID
func GetBrandByID(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		brandId := c.Param("id")
		objectId, err := primitive.ObjectIDFromHex(brandId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid brand ID"})
			return
		}

		var brand Brand
		err = app.BrandCollection.FindOne(c, bson.M{"_id": objectId}).Decode(&brand)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "Cannot find brand with the id: " + brandId})
			return
		}

		c.JSON(http.StatusOK, gin.H{"brand": brand})
	}
}

func ListSelectBrands(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var brands []Brand

		// Fetch brands for the specific merchant
		cursor, err := app.BrandCollection.Find(c, bson.M{"name": 1})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}
		defer cursor.Close(c)

		if err = cursor.All(c, &brands); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}

		c.JSON(http.StatusOK, gin.H{"brands": brands})
	}
}

func UpdateBrand(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		brandId := c.Param("id")
		objectId, err := primitive.ObjectIDFromHex(brandId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid brand ID"})
			return
		}

		var updateBrand BrandUpdate
		if err := c.ShouldBindJSON(&updateBrand); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		slug := updateBrand.Slug
		var foundBrand Brand
		err = app.BrandCollection.FindOne(c, bson.M{"slug": slug}).Decode(&foundBrand)
		if err == nil && foundBrand.ID != objectId {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Slug is already in use."})
			return
		}

		update := bson.M{
			"$set": updateBrand,
		}

		_, err = app.BrandCollection.UpdateOne(c, bson.M{"_id": objectId}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Brand has been updated successfully!",
		})
	}
}

func UpdateBrandActive(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		brandId := c.Param("id")
		objectId, err := primitive.ObjectIDFromHex(brandId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid brand ID"})
			return
		}

		var updateBrand struct {
			IsActive bool `json:"isActive"`
		}
		if err := c.ShouldBindJSON(&updateBrand); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		update := bson.M{
			"$set": bson.M{"isActive": updateBrand.IsActive},
		}

		_, err = app.BrandCollection.UpdateOne(c, bson.M{"_id": objectId}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Brand has been updated successfully!",
		})
	}
}

func DeleteBrand(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		brandId := c.Param("id")
		objectId, err := primitive.ObjectIDFromHex(brandId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid brand ID"})
			return
		}

		result, err := app.BrandCollection.DeleteOne(c, bson.M{"_id": objectId})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Your request could not be processed. Please try again."})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Brand has been deleted successfully!",
			"brand":   result,
		})
	}
}
