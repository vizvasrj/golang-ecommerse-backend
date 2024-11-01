package user

import (
	"src/pkg/module/merchant"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserRole type
type UserRole string

const (
	RoleAdmin    UserRole = "ROLE ADMIN"
	RoleMember   UserRole = "ROLE MEMBER"
	RoleMerchant UserRole = "ROLE MERCHANT"
)

type EmailProvider string

const (
	EmailProviderEmail    EmailProvider = "Email"
	EmailProviderGoogle   EmailProvider = "Google"
	EmailProviderFacebook EmailProvider = "Facebook"
)

type User struct {
	ID                   primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Email                string             `bson:"email,omitempty" json:"email,omitempty"`
	PhoneNumber          string             `bson:"phoneNumber,omitempty" json:"phoneNumber,omitempty"`
	FirstName            string             `bson:"firstName,omitempty" json:"firstName,omitempty"`
	LastName             string             `bson:"lastName,omitempty" json:"lastName,omitempty"`
	Password             string             `bson:"password,omitempty" json:"password,omitempty"`
	Merchant             primitive.ObjectID `bson:"merchant,omitempty" json:"merchant,omitempty"`
	Provider             EmailProvider      `bson:"provider,omitempty" json:"provider,omitempty"`
	GoogleID             string             `bson:"googleId,omitempty" json:"googleId,omitempty"`
	FacebookID           string             `bson:"facebookId,omitempty" json:"facebookId,omitempty"`
	Avatar               string             `bson:"avatar,omitempty" json:"avatar,omitempty"`
	Role                 UserRole           `bson:"role,omitempty" json:"role,omitempty"`
	ResetPasswordToken   string             `bson:"resetPasswordToken,omitempty" json:"resetPasswordToken,omitempty"`
	ResetPasswordExpires time.Time          `bson:"resetPasswordExpires,omitempty" json:"resetPasswordExpires,omitempty"`
	Updated              time.Time          `bson:"updated,omitempty" json:"updated,omitempty"`
	Created              time.Time          `bson:"created,omitempty" json:"created,omitempty"`
}

type UserSearch struct {
	ID                   primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Email                string             `bson:"email,omitempty" json:"email,omitempty"`
	PhoneNumber          string             `bson:"phoneNumber,omitempty" json:"phoneNumber,omitempty"`
	FirstName            string             `bson:"firstName,omitempty" json:"firstName,omitempty"`
	LastName             string             `bson:"lastName,omitempty" json:"lastName,omitempty"`
	Password             string             `bson:"password,omitempty" json:"password,omitempty"`
	Merchant             merchant.Merchant  `bson:"merchant,omitempty" json:"merchant,omitempty"`
	Provider             EmailProvider      `bson:"provider,omitempty" json:"provider,omitempty"`
	GoogleID             string             `bson:"googleId,omitempty" json:"googleId,omitempty"`
	FacebookID           string             `bson:"facebookId,omitempty" json:"facebookId,omitempty"`
	Avatar               string             `bson:"avatar,omitempty" json:"avatar,omitempty"`
	Role                 UserRole           `bson:"role,omitempty" json:"role,omitempty"`
	ResetPasswordToken   string             `bson:"resetPasswordToken,omitempty" json:"resetPasswordToken,omitempty"`
	ResetPasswordExpires time.Time          `bson:"resetPasswordExpires,omitempty" json:"resetPasswordExpires,omitempty"`
	Updated              time.Time          `bson:"updated,omitempty" json:"updated,omitempty"`
	Created              time.Time          `bson:"created,omitempty" json:"created,omitempty"`
}

type UserUpdate struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
}
