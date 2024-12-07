package main

import (
	"log"
	"os"
	"src/pkg/conf"
	"src/pkg/db"
	"src/pkg/env"
	"src/pkg/module/address"
	"src/pkg/module/auth"
	"src/pkg/module/brand"
	"src/pkg/module/cart"
	"src/pkg/module/category"
	"src/pkg/module/order"
	"src/pkg/module/payment"
	"src/pkg/module/product"
	"src/pkg/module/review"
	"src/pkg/module/user"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	envs, err := env.GetEnv()
	if err != nil {
		log.Fatalln(err)
	}

	clinet, err := db.InitializeMongoDB(envs.DBUri)
	if err != nil {
		log.Fatalln(err)
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
	ReceiptCollection := db.GetCollection(clinet, envs.DBName, "receipts")
	config := &conf.Config{
		// ContactCollection:  ContactCollection,
		DB:                 clinet,
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
		ReceiptCollection:  ReceiptCollection,

		Env:           envs,
		TokenLifetime: 24,
		MongoClient:   clinet,
	}

	// Start the server
	router := gin.Default()
	router.Use(normalizeURLMiddleware())
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r := router.Group("/api")
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message":   "pong",
			"clientUrl": os.Getenv("CLIENT_URL"),
		})
	})
	{

		auth.SetupRouter("/auth", r, config)
		address.SetupRouter("/address", r, config)
		brand.SetupRouter("/brand", r, config)
		product.SetupRouter("/product", r, config)
		user.SetupRouter("/user", r, config)
		user.SetupRouter_Merchant("/merchant", r, config)
		category.SetupRoute("/category", r, config)
		cart.SetupRoute("/cart", r, config)
		order.SetupRoute("/order", r, config)
		review.SetupRouter("/review", r, config)
		payment.SetupRouter("/payment", r, config)
	}

	router.Run(":3000")

}

// normalizeURLMiddleware ensures that all URLs are treated consistently
func normalizeURLMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		// Remove trailing slash if it's not the root path
		if path != "/" && strings.HasSuffix(path, "/") {
			c.Request.URL.Path = strings.TrimSuffix(path, "/")
		}
		c.Next()
	}
}
