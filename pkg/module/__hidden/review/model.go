package review

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ReviewStatus type
type ReviewStatus string

const (
	Rejected        ReviewStatus = "Rejected"
	Approved        ReviewStatus = "Approved"
	WaitingApproval ReviewStatus = "Waiting Approval"
)

type ReviewUser struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	FirstName string             `bson:"firstName,omitempty" json:"firstName,omitempty"`
	LastName  string             `bson:"lastName,omitempty" json:"lastName,omitempty"`
	Email     string             `bson:"email,omitempty" json:"email,omitempty"`
}

type Review struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	Product       primitive.ObjectID `bson:"product,omitempty" json:"product,omitempty"`
	User          ReviewUser         `bson:"user,omitempty" json:"user,omitempty"`
	Title         string             `bson:"title,omitempty" json:"title,omitempty"`
	Rating        float64            `bson:"rating,omitempty" json:"rating,omitempty"`
	Review        string             `bson:"review,omitempty" json:"review,omitempty"`
	IsRecommended bool               `bson:"isRecommended,omitempty" json:"isRecommended,omitempty"`
	Status        ReviewStatus       `bson:"status,omitempty" json:"status,omitempty"`
	Updated       time.Time          `bson:"updated,omitempty" json:"updated,omitempty"`
	Created       time.Time          `bson:"created,omitempty" json:"created,omitempty"`
}

type PutReviewInput struct {
	Product       primitive.ObjectID `bson:"product,omitempty" json:"product,omitempty"`
	Title         string             `bson:"title" json:"title" binding:"required"`
	Rating        string             `bson:"rating" json:"rating" binding:"required"`
	Review        string             `bson:"review" json:"review" binding:"required"`
	IsRecommended bool               `bson:"isRecommended" json:"isRecommended"`
	Updated       time.Time
}
