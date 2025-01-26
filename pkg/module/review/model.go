package review

import (
	"time"

	"github.com/google/uuid"
	"github.com/guregu/null/v5"
)

// ReviewStatus type
type ReviewStatus string

const (
	Rejected        ReviewStatus = "Rejected"
	Approved        ReviewStatus = "Approved"
	WaitingApproval ReviewStatus = "Waiting Approval"
)

// type ReviewUser struct {
// 	ID        primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
// 	FirstName string             `bson:"firstName,omitempty" json:"firstName,omitempty"`
// 	LastName  string             `bson:"lastName,omitempty" json:"lastName,omitempty"`
// 	Email     string             `bson:"email,omitempty" json:"email,omitempty"`
// }

// type Review struct {
// 	ID            primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
// 	Product       primitive.ObjectID `bson:"product,omitempty" json:"product,omitempty"`
// 	User          ReviewUser         `bson:"user,omitempty" json:"user,omitempty"`
// 	Title         string             `bson:"title,omitempty" json:"title,omitempty"`
// 	Rating        float64            `bson:"rating,omitempty" json:"rating,omitempty"`
// 	Review        string             `bson:"review,omitempty" json:"review,omitempty"`
// 	IsRecommended bool               `bson:"isRecommended,omitempty" json:"isRecommended,omitempty"`
// 	Status        ReviewStatus       `bson:"status,omitempty" json:"status,omitempty"`
// 	Updated       time.Time          `bson:"updated,omitempty" json:"updated,omitempty"`
// 	Created       time.Time          `bson:"created,omitempty" json:"created,omitempty"`
// }

// type PutReviewInput struct {
// 	Product       primitive.ObjectID `bson:"product,omitempty" json:"product,omitempty"`
// 	Title         string             `bson:"title" json:"title" binding:"required"`
// 	Rating        string             `bson:"rating" json:"rating" binding:"required"`
// 	Review        string             `bson:"review" json:"review" binding:"required"`
// 	IsRecommended bool               `bson:"isRecommended" json:"isRecommended"`
// 	Updated       time.Time
// }

type Review struct {
	ID            uuid.UUID   `db:"id" json:"id"`
	ProductID     uuid.UUID   `db:"product_id" json:"product_id"`
	UserID        uuid.UUID   `db:"user_id" json:"user_id"`
	Title         string      `db:"title" json:"title" binding:"required"`
	Rating        float64     `db:"rating" json:"rating" binding:"required"`
	Review        string      `db:"review" json:"review" binding:"required"`
	IsRecommended bool        `db:"is_recommended" json:"is_recommended"`
	Status        string      `db:"status" json:"status"`
	Updated       null.Time   `db:"updated" json:"updated"`
	Created       time.Time   `db:"created" json:"created"`
	User          *ReviewUser `json:"user,omitempty"`
}

type ReviewUser struct { // This struct represents the user who wrote the review.
	ID        uuid.UUID `db:"id" json:"id"`
	FirstName string    `db:"first_name" json:"first_name"`
	LastName  string    `db:"last_name" json:"last_name"`
	Email     string    `db:"email" json:"email"`
}

type PutReviewInput struct {
	ProductID     uuid.UUID `json:"product_id" binding:"required"`
	Title         string    `json:"title" binding:"required"`
	Rating        string    `json:"rating" binding:"required"`
	Review        string    `json:"review" binding:"required"`
	IsRecommended bool      `json:"is_recommended"`
}
