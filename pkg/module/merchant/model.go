package merchant

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MerchantStatus type
type MerchantStatus string

// MerchantStatus constants
const (
	WaitingApproval MerchantStatus = "Waiting Approval"
	Rejected        MerchantStatus = "Rejected"
	Approved        MerchantStatus = "Approved"
)

// Merchant struct
type Merchant struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Email       string             `bson:"email" json:"email"`
	PhoneNumber string             `bson:"phoneNumber" json:"phoneNumber"`
	BrandName   string             `bson:"brandName" json:"brandName"`
	Business    string             `bson:"business" json:"business"`
	IsActive    bool               `bson:"isActive" json:"isActive"`
	Brand       string             `bson:"brand" json:"brand"`
	Status      MerchantStatus     `bson:"status" json:"status"`
	Updated     time.Time          `bson:"updated" json:"updated"`
	Created     time.Time          `bson:"created" json:"created"`
}

type MerchantAdd struct {
	Name        string `form:"name" binding:"required"`
	Email       string `form:"email" binding:"required"`
	PhoneNumber string `form:"phoneNumber" binding:"required"`
	BrandName   string `form:"brandName" binding:"required"`
	Business    string `form:"business" binding:"required"`
}
