package contact

import "time"

// Contact model
type Contact struct {
	ID      uint      `json:"id"`
	Name    string    `json:"name"`
	Email   string    `json:"email"`
	Message string    `json:"message"`
	Updated time.Time `json:"updated"`
	Created time.Time `json:"created"`
}
