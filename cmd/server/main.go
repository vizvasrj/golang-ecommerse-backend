package main

import (
	"src/pkg/conf"
	"src/pkg/db"
	"src/pkg/env"
	"src/pkg/module/address"
	"src/pkg/module/auth"

	"github.com/gin-gonic/gin"
)

func main() {
	envs, err := env.GetEnv()
	if err != nil {
		panic(err)
	}

	clinet, err := db.InitializeMongoDB(envs.DBUri)
	if err != nil {
		panic(err)
	}

	AddressCollection := db.GetCollection(clinet, envs.DBName, "addresses")
	CartCollection := db.GetCollection(clinet, envs.DBName, "carts")
	// ContactCollection := db.GetCollection(clinet, envs.DBName, "contact")
	WishlistCollection := db.GetCollection(clinet, envs.DBName, "wishlists")
	BrandCollection := db.GetCollection(clinet, envs.DBName, "brands")
	ProductCollection := db.GetCollection(clinet, envs.DBName, "products")
	OrderCollection := db.GetCollection(clinet, envs.DBName, "orders")
	ReviewCollection := db.GetCollection(clinet, envs.DBName, "reviews")
	UserCollection := db.GetCollection(clinet, envs.DBName, "users")
	MerchantCollection := db.GetCollection(clinet, envs.DBName, "merchants")
	CategoryCollection := db.GetCollection(clinet, envs.DBName, "categories")
	config := &conf.Config{
		// ContactCollection:  ContactCollection,
		AddressCollection:  AddressCollection,
		CartCollection:     CartCollection,
		WishlistCollection: WishlistCollection,
		BrandCollection:    BrandCollection,
		ProductCollection:  ProductCollection,
		OrderCollection:    OrderCollection,
		ReviewCollection:   ReviewCollection,
		UserCollection:     UserCollection,
		MerchantCollection: MerchantCollection,
		CategoryCollection: CategoryCollection,

		Env:           envs,
		TokenLifetime: 24,
	}

	// Start the server
	r := gin.Default()
	auth.SetupRouter("/auth", r, config)
	address.SetupRouter("/address", r, config)
	r.Run(":3000")

}
