package merchant

import "time"

// MerchantStatus type
type MerchantStatus string

// MerchantStatus constants
const (
	WaitingApproval MerchantStatus = "Waiting Approval"
	Rejected        MerchantStatus = "Rejected"
	Approved        MerchantStatus = "Approved"
)

// Merchant model
type Merchant struct {
	ID          uint           `json:"id"`
	Name        string         `json:"name"`
	Email       string         `json:"email"`
	PhoneNumber string         `json:"phoneNumber"`
	BrandName   string         `json:"brandName"`
	Business    string         `json:"business"`
	IsActive    bool           `json:"isActive"`
	BrandID     uint           `json:"brandId"`
	Status      MerchantStatus `json:"status"`
	Updated     time.Time      `json:"updated"`
	Created     time.Time      `json:"created"`
}
