package common

import (
	"time"

	"github.com/google/uuid"
	"github.com/guregu/null/v5"
)

type EmailProvider string

const (
	EmailProviderEmail    EmailProvider = "Email"
	EmailProviderGoogle   EmailProvider = "Google"
	EmailProviderFacebook EmailProvider = "Facebook"
)

type User struct {
	ID                   uuid.UUID     `db:"id" json:"_id"`
	Email                string        `db:"email" json:"email"`
	PhoneNumber          null.String   `db:"phone_number" json:"phoneNumber"`
	FirstName            string        `db:"first_name" json:"firstName"`
	LastName             string        `db:"last_name" json:"lastName"`
	Password             string        `db:"password" json:"-"`
	MerchantID           uuid.NullUUID `db:"merchant_id" json:"merchantId"`
	Provider             EmailProvider `db:"provider" json:"provider"`
	GoogleID             null.String   `db:"google_id" json:"googleId"`
	FacebookID           null.String   `db:"facebook_id" json:"facebookId"`
	Avatar               null.String   `db:"avatar" json:"avatar"`
	Role                 string        `db:"role" json:"role"`
	ResetPasswordToken   null.String   `db:"reset_password_token" json:"-"`
	ResetPasswordExpires null.Time     `db:"reset_password_expires" json:"-"`
	Updated              null.Time     `db:"updated_at" json:"updatedAt"`
	Created              time.Time     `db:"created_at" json:"createdAt"`
}

type MerchantStatus string

// MerchantStatus constants
const (
	WaitingApproval MerchantStatus = "Waiting Approval"
	Rejected        MerchantStatus = "Rejected"
	Approved        MerchantStatus = "Approved"
)

type Merchant struct {
	ID          uuid.UUID      `db:"id" json:"Id"`
	UserID      uuid.UUID      `db:"user_id" json:"userId"`
	Name        string         `db:"name" json:"name"`
	Email       string         `db:"email" json:"email"`
	PhoneNumber string         `db:"phone_number" json:"phoneNumber"`
	BrandName   string         `db:"brand_name" json:"brandName"`
	Business    string         `db:"business" json:"business"`
	IsActive    bool           `db:"is_active" json:"isActive"`
	BrandID     uuid.NullUUID  `db:"brand_id" json:"brandId"`
	Status      MerchantStatus `db:"status" json:"status"`
	Updated     null.Time      `db:"updated" json:"updated"`
	Created     time.Time      `db:"created" json:"created"`
}
