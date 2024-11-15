package product

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Brand struct {
	ID   primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name string             `bson:"name" json:"name"`
}

type Product struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	SKU         string             `bson:"sku,omitempty" json:"sku,omitempty"`
	Name        string             `bson:"name,omitempty" json:"name,omitempty"`
	Slug        string             `bson:"slug,omitempty" json:"slug,omitempty"`
	ImageURL    string             `bson:"imageUrl,omitempty" json:"imageUrl,omitempty"`
	ImageKey    string             `bson:"imageKey,omitempty" json:"imageKey,omitempty"`
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	Quantity    int                `bson:"quantity,omitempty" json:"quantity,omitempty"`
	Price       float64            `bson:"price,omitempty" json:"price,omitempty"`
	Taxable     bool               `bson:"taxable,omitempty" json:"taxable,omitempty"`
	IsActive    bool               `bson:"isActive,omitempty" json:"isActive,omitempty"`
	Brand       Brand              `bson:"brand,omitempty" json:"brand,omitempty"`
	Updated     time.Time          `bson:"updated,omitempty" json:"updated,omitempty"`
	Created     time.Time          `bson:"created,omitempty" json:"created,omitempty"`
	Merchant    primitive.ObjectID `bson:"merchant,omitempty" json:"merchant,omitempty"`
}

type AddProductInput struct {
	SKU         string  `form:"sku" binding:"required"`
	Name        string  `form:"name" binding:"required"`
	Description string  `form:"description" binding:"required"`
	Quantity    int     `form:"quantity" binding:"required"`
	Price       float64 `form:"price" binding:"required"`
	Taxable     bool    `form:"taxable"`
	IsActive    bool    `form:"isActive"`
	Brand       string  `form:"brand"`
}

type IndividualProduct struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	SKU         string             `bson:"sku,omitempty" json:"sku,omitempty"`
	Name        string             `bson:"name,omitempty" json:"name,omitempty"`
	Slug        string             `bson:"slug,omitempty" json:"slug,omitempty"`
	ImageURL    string             `bson:"imageUrl,omitempty" json:"imageUrl,omitempty"`
	ImageKey    string             `bson:"imageKey,omitempty" json:"imageKey,omitempty"`
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	Quantity    int                `bson:"quantity,omitempty" json:"quantity,omitempty"`
	Price       float64            `bson:"price,omitempty" json:"price,omitempty"`
	Taxable     bool               `bson:"taxable,omitempty" json:"taxable,omitempty"`
	IsActive    bool               `bson:"isActive,omitempty" json:"isActive,omitempty"`
	Brand       primitive.ObjectID `bson:"brand,omitempty" json:"brand,omitempty"`
	Updated     time.Time          `bson:"updated,omitempty" json:"updated,omitempty"`
	Created     time.Time          `bson:"created,omitempty" json:"created,omitempty"`
	Merchant    primitive.ObjectID `bson:"merchant,omitempty" json:"merchant,omitempty"`
}
