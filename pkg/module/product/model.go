package product

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Product model
type Product struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Sku         string    `gorm:"column:sku;type:varchar(255)" json:"sku"`
	Name        string    `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Slug        string    `gorm:"column:slug;type:varchar(255);unique;not null" json:"slug"`
	ImageUrl    string    `gorm:"column:image_url;type:varchar(255)" json:"imageUrl"`
	ImageKey    string    `gorm:"column:image_key;type:varchar(255)" json:"imageKey"`
	Description string    `gorm:"column:description;type:text" json:"description"`
	Quantity    int       `gorm:"column:quantity;not null" json:"quantity"`
	Price       float64   `gorm:"column:price;not null" json:"price"`
	Taxable     bool      `gorm:"column:taxable;default:false" json:"taxable"`
	IsActive    bool      `gorm:"column:is_active;default:true" json:"isActive"`
	BrandID     uint      `gorm:"column:brand_id;default:null" json:"brand"`
	Updated     time.Time `gorm:"column:updated" json:"updated"`
	Created     time.Time `gorm:"column:created;default:CURRENT_TIMESTAMP" json:"created"`
	MerchantID  uint      `gorm:"column:merchant_id;default:null" json:"merchant"`
}

// TableName specifies the table name for the Product model
func (Product) TableName() string {
	return "products"
}

// CreateProduct creates a new product record in the database
func CreateProduct(db *gorm.DB, product *Product) error {
	// Generate a slug from the product name
	slug := strings.ToLower(strings.ReplaceAll(product.Name, " ", "-"))

	// Generate a random 7-character hexadecimal string
	randomHex, err := GenerateRandomHex(4) // 4 bytes = 8 hex characters
	if err != nil {
		return err
	}

	// Append the random hex to the slug
	product.Slug = fmt.Sprintf("%s-%s", slug, randomHex[:7])

	// Create the product in the database
	return db.Create(product).Error
}

// * move to misc.go
// GenerateRandomHex generates a random 7-character hexadecimal string
func GenerateRandomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
