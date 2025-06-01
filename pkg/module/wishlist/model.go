package wishlist

import (
	"time"

	"github.com/google/uuid"
)

type Wishlist struct {
	ID      uuid.UUID `bson:"_id,omitempty" json:"_id,omitempty"`
	Product uuid.UUID `bson:"product,omitempty" json:"product,omitempty"`
	User    uuid.UUID `bson:"user,omitempty" json:"user,omitempty"`
	IsLiked bool      `bson:"isLiked,omitempty" json:"isLiked,omitempty"`
	Updated time.Time `bson:"updated,omitempty" json:"updated,omitempty"`
	Created time.Time `bson:"created,omitempty" json:"created,omitempty"`
}
