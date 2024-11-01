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

type Review struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Product       primitive.ObjectID `bson:"product,omitempty" json:"product,omitempty"`
	User          primitive.ObjectID `bson:"user,omitempty" json:"user,omitempty"`
	Title         string             `bson:"title,omitempty" json:"title,omitempty"`
	Rating        float64            `bson:"rating,omitempty" json:"rating,omitempty"`
	Review        string             `bson:"review,omitempty" json:"review,omitempty"`
	IsRecommended bool               `bson:"isRecommended,omitempty" json:"isRecommended,omitempty"`
	Status        ReviewStatus       `bson:"status,omitempty" json:"status,omitempty"`
	Updated       time.Time          `bson:"updated,omitempty" json:"updated,omitempty"`
	Created       time.Time          `bson:"created,omitempty" json:"created,omitempty"`
}
