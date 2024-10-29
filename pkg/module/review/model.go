package review

import (
	"database/sql/driver"
	"time"

	_ "gorm.io/gorm"
)

// ReviewStatus type
type ReviewStatus string

const (
	Rejected        ReviewStatus = "Rejected"
	Approved        ReviewStatus = "Approved"
	WaitingApproval ReviewStatus = "Waiting Approval"
)

// Scan implements the Scanner interface for ReviewStatus
func (rs *ReviewStatus) Scan(value interface{}) error {
	*rs = ReviewStatus(value.([]byte))
	return nil
}

// Value implements the Valuer interface for ReviewStatus
func (rs ReviewStatus) Value() (driver.Value, error) {
	return string(rs), nil
}

// Review model
type Review struct {
	ID            uint         `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductID     uint         `gorm:"column:product_id;default:null" json:"product"`
	UserID        uint         `gorm:"column:user_id;default:null" json:"user"`
	Title         string       `gorm:"column:title;type:varchar(255)" json:"title"`
	Rating        int          `gorm:"column:rating;default:0" json:"rating"`
	Review        string       `gorm:"column:review;type:text" json:"review"`
	IsRecommended bool         `gorm:"column:is_recommended;default:true" json:"isRecommended"`
	Status        ReviewStatus `gorm:"column:status;type:varchar(255);default:'Waiting Approval'" json:"status"`
	Updated       time.Time    `gorm:"column:updated" json:"updated"`
	Created       time.Time    `gorm:"column:created;default:CURRENT_TIMESTAMP" json:"created"`
}

// TableName specifies the table name for the Review model
func (Review) TableName() string {
	return "reviews"
}
