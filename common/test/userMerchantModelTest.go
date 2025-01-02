package main

import (
	"encoding/json"
	"fmt"
	"src/common"
	"time"

	"github.com/google/uuid"
	"github.com/guregu/null/v5"
)

type User struct {
	ID                   uuid.UUID     `db:"id" json:"id"`
	Email                string        `db:"email" json:"email"`
	PhoneNumber          null.String   `db:"phone_number" json:"phone_number"`
	FirstName            string        `db:"first_name" json:"first_name"`
	LastName             string        `db:"last_name" json:"last_name"`
	Password             string        `db:"password" json:"-"`
	MerchantID           uuid.NullUUID `db:"merchant_id" json:"merchant_id"`
	Provider             null.String   `db:"provider" json:"provider"`
	GoogleID             null.String   `db:"google_id" json:"google_id"`
	FacebookID           null.String   `db:"facebook_id" json:"facebook_id"`
	Avatar               null.String   `db:"avatar" json:"avatar"`
	Role                 string        `db:"role" json:"role"`
	ResetPasswordToken   null.String   `db:"reset_password_token" json:"-"`
	ResetPasswordExpires null.Time     `db:"reset_password_expires" json:"-"`
	Updated              null.Time     `db:"updated_at" json:"updated_at"`
	Created              time.Time     `db:"created_at" json:"created_at"`
}

func main() {
	sampleUser := common.User{
		ID:          uuid.New(),
		Email:       "user@example.com",
		PhoneNumber: null.NewString("1234567890", true),
		FirstName:   "John",
		LastName:    "Doe",
		Password:    "secret", // This will not be included in JSON output
		MerchantID:  uuid.NullUUID{UUID: uuid.New(), Valid: true},
		Provider:    common.EmailProviderEmail,
		GoogleID:    null.StringFrom("google123"),
		FacebookID:  null.StringFrom("facebook123"),
		Avatar:      null.StringFrom("https://example.com/avatar.jpg"),
		Role:        "Member",
	}

	// Convert the User struct to JSON
	jsonData, err := json.MarshalIndent(sampleUser, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling to JSON:", err)
		return
	}

	// Print the indented JSON
	fmt.Println(string(jsonData))

	jsonString := `{
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "email": "user@example.com",
        "phone_number": "1234567890",
        "first_name": "John",
        "last_name": "Doe",
        "merchant_id": "550e8400-e29b-41d4-a716-446655440001",
        "provider": "Email",
        "google_id": "google123",
        "facebook_id": "facebook123",
        "avatar": "https://example.com/avatar.jpg",
        "role": "Member",
        "updated_at": null,
        "created_at": "2023-10-01T10:00:00Z"
    }`

	// Unmarshal the JSON string into a User struct
	var user User
	err = json.Unmarshal([]byte(jsonString), &user)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return
	}

	// Print the User struct
	fmt.Printf("Unmarshaled User struct: %+v\n", user)

}
