package wishlist

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Wishlist struct {
	ID      primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Product primitive.ObjectID `bson:"product,omitempty" json:"product,omitempty"`
	User    primitive.ObjectID `bson:"user,omitempty" json:"user,omitempty"`
	IsLiked bool               `bson:"isLiked,omitempty" json:"isLiked,omitempty"`
	Updated time.Time          `bson:"updated,omitempty" json:"updated,omitempty"`
	Created time.Time          `bson:"created,omitempty" json:"created,omitempty"`
}
