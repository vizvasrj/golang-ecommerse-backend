package user

import (
	"database/sql/driver"
	"errors"
	"time"
)

// UserRole type
type UserRole string

const (
	RoleAdmin    UserRole = "ROLE ADMIN"
	RoleMember   UserRole = "ROLE MEMBER"
	RoleMerchant UserRole = "ROLE MERCHANT"
)

// Scan implements the Scanner interface for UserRole
func (ur *UserRole) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		*ur = UserRole(v)
	case string:
		*ur = UserRole(v)
	default:
		return errors.New("invalid type for UserRole")
	}
	return nil
}

// Value implements the Valuer interface for UserRole
func (ur UserRole) Value() (driver.Value, error) {
	return string(ur), nil
}

// User model
type User struct {
	ID                   uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Email                string    `gorm:"column:email" json:"email"`
	PhoneNumber          string    `gorm:"column:phoneNumber" json:"phoneNumber"`
	FirstName            string    `gorm:"column:firstName" json:"firstName"`
	LastName             string    `gorm:"column:lastName" json:"lastName"`
	Password             string    `gorm:"column:password" json:"password"`
	MerchantID           uint      `gorm:"column:merchant;default:null" json:"merchant"`
	Provider             string    `gorm:"column:provider;default:'email'" json:"provider"`
	GoogleID             string    `gorm:"column:googleId" json:"googleId"`
	FacebookID           string    `gorm:"column:facebookId" json:"facebookId"`
	Avatar               string    `gorm:"column:avatar" json:"avatar"`
	Role                 UserRole  `gorm:"column:role;type:varchar(255);default:'ROLE MEMBER'" json:"role"`
	ResetPasswordToken   string    `gorm:"column:resetPasswordToken" json:"resetPasswordToken"`
	ResetPasswordExpires time.Time `gorm:"column:resetPasswordExpires" json:"resetPasswordExpires"`
	Updated              time.Time `gorm:"column:updated" json:"updated"`
	Created              time.Time `gorm:"column:created;default:CURRENT_TIMESTAMP" json:"created"`
}

// TableName specifies the table name for the User model
func (User) TableName() string {
	return "users"
}
