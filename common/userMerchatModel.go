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
	ID                   uuid.UUID     `db:"id" json:"id"`
	Email                string        `db:"email" json:"email"`
	PhoneNumber          null.String   `db:"phone_number" json:"phone_number"`
	FirstName            string        `db:"first_name" json:"first_name"`
	LastName             string        `db:"last_name" json:"last_name"`
	Password             string        `db:"password" json:"-"`
	MerchantID           uuid.NullUUID `db:"merchant_id" json:"merchant_id"`
	Provider             EmailProvider `db:"provider" json:"provider"`
	GoogleID             null.String   `db:"google_id" json:"google_id"`
	FacebookID           null.String   `db:"facebook_id" json:"facebook_id"`
	Avatar               null.String   `db:"avatar" json:"avatar"`
	Role                 string        `db:"role" json:"role"`
	ResetPasswordToken   null.String   `db:"reset_password_token" json:"-"`
	ResetPasswordExpires null.Time     `db:"reset_password_expires" json:"-"`
	Updated              null.Time     `db:"updated_at" json:"updated_at"`
	Created              time.Time     `db:"created_at" json:"created_at"`
}

type MerchantStatus string

// MerchantStatus constants
const (
	WaitingApproval MerchantStatus = "Waiting Approval"
	Rejected        MerchantStatus = "Rejected"
	Approved        MerchantStatus = "Approved"
)

type Merchant struct {
	ID          uuid.UUID      `db:"id" json:"id"`
	UserID      uuid.UUID      `db:"user_id" json:"user_id"`
	Name        string         `db:"name" json:"name"`
	Email       string         `db:"email" json:"email"`
	PhoneNumber string         `db:"phone_number" json:"phone_number"`
	BrandName   string         `db:"brand_name" json:"brand_name"`
	Business    string         `db:"business" json:"business"`
	IsActive    bool           `db:"is_active" json:"is_active"`
	BrandID     uuid.NullUUID  `db:"brand_id" json:"brand_id"`
	Status      MerchantStatus `db:"status" json:"status"`
	Updated     null.Time      `db:"updated" json:"updated"`
	Created     time.Time      `db:"created" json:"created"`
}
