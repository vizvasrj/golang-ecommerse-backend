package conf

import (
	"src/pkg/env"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

type Config struct {
	DB                 *mongo.Client
	AddressCollection  *mongo.Collection
	CartCollection     *mongo.Collection
	ContactCollection  *mongo.Collection
	WishlistCollection *mongo.Collection
	BrandCollection    *mongo.Collection
	ProductCollection  *mongo.Collection
	OrderCollection    *mongo.Collection
	ReviewCollection   *mongo.Collection
	UserCollection     *mongo.Collection
	MerchantCollection *mongo.Collection
	CategoryCollection *mongo.Collection
	ReceiptCollection  *mongo.Collection
	Env                *env.Env
	TokenLifetime      time.Duration
	MongoClient        *mongo.Client
}
