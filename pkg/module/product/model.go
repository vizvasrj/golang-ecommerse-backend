package product

import (
	"src/pkg/module/brand"
	category "src/pkg/module/category"
	"time"

	"github.com/google/uuid"
	"github.com/guregu/null/v5"
)

type Product struct {
	ID          uuid.UUID     `db:"id" json:"id"`
	SKU         string        `db:"sku" json:"sku" binding:"required"`
	Name        string        `db:"name" json:"name" binding:"required"`
	Slug        string        `db:"slug" json:"slug"`
	ImageURL    null.String   `db:"image_url" json:"image_url"`
	ImageKey    null.String   `db:"image_key" json:"image_key"`
	Description string        `db:"description" json:"description" binding:"required"`
	Quantity    int           `db:"quantity" json:"quantity" binding:"required"`
	Price       float64       `db:"price" json:"price" binding:"required"`
	Taxable     bool          `db:"taxable" json:"taxable"`
	IsActive    bool          `db:"is_active" json:"is_active"`
	BrandID     uuid.NullUUID `db:"brand_id" json:"brand_id"`
	CategoryID  uuid.NullUUID `db:"category_id" json:"category_id"`
	MerchantID  uuid.UUID     `db:"merchant_id" json:"merchant_id"`
	Updated     null.Time     `db:"updated" json:"updated"`
	Created     time.Time     `db:"created" json:"created"`
}

type ProductUpdate struct { // Struct for partial updates
	SKU         *string    `json:"sku"`
	Name        *string    `json:"name"`
	Slug        *string    `json:"slug"`
	ImageURL    *string    `json:"image_url"`
	ImageKey    *string    `json:"image_key"`
	Description *string    `json:"description"`
	Quantity    *int       `json:"quantity"`
	Price       *float64   `json:"price"`
	Taxable     *bool      `json:"taxable"`
	IsActive    *bool      `json:"is_active"`
	BrandID     *uuid.UUID `json:"brand_id"`
	CategoryID  *uuid.UUID `json:"category_id"`
}

type AddProductInput struct { // Request input struct for AddProduct
	SKU         string    `form:"sku" binding:"required"`
	Name        string    `form:"name" binding:"required"`
	Slug        string    `form:"slug" binding:"required"`
	Description string    `form:"description" binding:"required"`
	Quantity    int       `form:"quantity" binding:"required"`
	Price       float64   `form:"price" binding:"required"`
	Taxable     bool      `form:"taxable"`
	IsActive    bool      `form:"is_active"`
	BrandID     uuid.UUID `form:"brand_id"`
	CategoryID  uuid.UUID `form:"category_id"`
}

type GetProduct struct {
	ID          uuid.UUID           `json:"id"`
	SKU         string              `json:"sku"`
	Name        string              `json:"name"`
	Slug        string              `json:"slug"`
	ImageURL    []string            `json:"image_url"`
	Description string              `json:"description"`
	Quantity    int                 `json:"quantity"`
	Price       float64             `json:"price"`
	Taxable     bool                `json:"taxable"`
	IsActive    bool                `json:"is_active"`
	Brand       brand.Brand         `json:"brand_id,omitempty"`
	Categories  []category.Category `json:"categories,omitempty"`
	MerchantID  uuid.UUID           `json:"merchant_id,omitempty"`
	Updated     time.Time           `json:"updated,omitempty"`
	Created     time.Time           `json:"created,omitempty"`
}
