package product

import (
	"src/pkg/module/brand"
	category "src/pkg/module/category"
	"time"

	"github.com/google/uuid"
	"github.com/guregu/null/v5"
)

type Product struct {
	ID          uuid.UUID     `db:"id" json:"_id"`
	SKU         string        `db:"sku" json:"sku" binding:"required"`
	Name        string        `db:"name" json:"name" binding:"required"`
	Slug        string        `db:"slug" json:"slug"`
	ImageURL    null.String   `db:"image_url" json:"imageUrl"`
	ImageKey    null.String   `db:"image_key" json:"imageKey"`
	Description string        `db:"description" json:"description" binding:"required"`
	Quantity    int           `db:"quantity" json:"quantity" binding:"required"`
	Price       float64       `db:"price" json:"price" binding:"required"`
	Taxable     bool          `db:"taxable" json:"taxable"`
	IsActive    bool          `db:"is_active" json:"isActive"`
	BrandID     uuid.NullUUID `db:"brand_id" json:"brandId"`
	CategoryID  uuid.NullUUID `db:"category_id" json:"categoryId"`
	MerchantID  uuid.UUID     `db:"merchant_id" json:"merchantId"`
	Updated     null.Time     `db:"updated" json:"updated"`
	Created     time.Time     `db:"created" json:"created"`
}

type ProductUpdate struct { // Struct for partial updates
	SKU         *string    `json:"sku"`
	Name        *string    `json:"name"`
	Slug        *string    `json:"slug"`
	ImageURL    *string    `json:"imageUrl"`
	ImageKey    *string    `json:"imageKey"`
	Description *string    `json:"description"`
	Quantity    *int       `json:"quantity"`
	Price       *float64   `json:"price"`
	Taxable     *bool      `json:"taxable"`
	IsActive    *bool      `json:"isActive"`
	BrandID     *uuid.UUID `json:"brandId"`
	CategoryID  *uuid.UUID `json:"categoryId"`
}

type AddProductInput struct { // Request input struct for AddProduct
	SKU         string    `form:"sku" binding:"required"`
	Name        string    `form:"name" binding:"required"`
	Slug        string    `form:"slug" binding:"required"`
	Description string    `form:"description" binding:"required"`
	Quantity    int       `form:"quantity" binding:"required"`
	Price       float64   `form:"price" binding:"required"`
	Taxable     bool      `form:"taxable"`
	IsActive    bool      `form:"isActive"`
	BrandID     uuid.UUID `form:"brandId"`
	CategoryID  uuid.UUID `form:"categoryId"`
}

type GetProduct struct {
	ID          uuid.UUID           `json:"_id"`
	SKU         string              `json:"sku"`
	Name        string              `json:"name"`
	Slug        string              `json:"slug"`
	ImageURL    []string            `json:"imageUrl"`
	Description string              `json:"description"`
	Quantity    int                 `json:"quantity"`
	Price       float64             `json:"price"`
	Taxable     bool                `json:"taxable"`
	IsActive    bool                `json:"isActive"`
	Brand       brand.Brand         `json:"brandId,omitempty"`
	Categories  []category.Category `json:"categories,omitempty"`
	MerchantID  uuid.UUID           `json:"merchantId,omitempty"`
	Updated     time.Time           `json:"updated,omitempty"`
	Created     time.Time           `json:"created,omitempty"`
}
