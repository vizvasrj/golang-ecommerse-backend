package main

import (
	"log"
	"src/pkg/module/address"
	"src/pkg/module/brand"
	"src/pkg/module/cart"
	"src/pkg/module/category"
	"src/pkg/module/contact"
	"src/pkg/module/merchant"
	"src/pkg/module/order"
	"src/pkg/module/product"
	"src/pkg/module/review"
	"src/pkg/module/user"
	"src/pkg/module/wishlist"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Initialize the database connection
	dsn := "host=localhost user=postgres password=postgres dbname=ecomm port=5432 sslmode=disable TimeZone=Asia/Kolkata"

	// Initialize the database connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database")
	}

	// err = db.Migrator().DropTable(&cart.Cart{})
	// if err != nil {
	// 	log.Fatalf("failed to drop Brand table: %v", err)
	// }
	// err = db.Migrator().DropTable(&cart.CartItem{})
	// if err != nil {
	// 	log.Fatalf("failed to drop Brand table: %v", err)
	// }

	// Migrate the schema
	err = db.AutoMigrate(
		&address.Address{},
		&brand.Brand{},
		&cart.Cart{},
		&cart.CartItem{},
		&category.Category{},
		&contact.Contact{},
		&merchant.Merchant{},
		&order.Order{},
		&product.Product{},
		&review.Review{},
		&user.User{},
		&wishlist.Wishlist{},
	)
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	log.Println("Database migration completed successfully")
}
