package wishlist

import (
	"src/pkg/module/product"
	"src/pkg/module/user"
	"time"
)

// Wishlist model
type Wishlist struct {
	ID        uint            `gorm:"primaryKey" json:"id"`
	ProductID uint            `gorm:"column:product;default:null" json:"product"`
	Product   product.Product `gorm:"foreignKey:ProductID" json:"-"`
	UserID    uint            `gorm:"column:user;default:null" json:"user"`
	User      user.User       `gorm:"foreignKey:UserID" json:"-"`
	IsLiked   bool            `gorm:"column:isLiked;default:false" json:"isLiked"`
	Updated   time.Time       `gorm:"column:updated;default:CURRENT_TIMESTAMP" json:"updated"`
	Created   time.Time       `gorm:"column:created;default:CURRENT_TIMESTAMP" json:"created"`
}

// TableName specifies the table name for the Wishlist model
func (Wishlist) TableName() string {
	return "wishlist"
}
