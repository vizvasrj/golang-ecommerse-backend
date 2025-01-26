package product

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"

	"src/common"
	"src/l"
	"src/pkg/conf"
	"src/pkg/misc"
	"src/pkg/module/brand"
	category "src/pkg/module/category"
)

func GetProductBySlug(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.Param("slug")

		var product Product // Use your Product struct
		query := `
		SELECT 
			id, sku, name, slug, image_url, image_key, description, quantity, price, taxable, is_active, brand_id, merchant_id, updated, created, merchant_id
		FROM products WHERE slug = $1 AND is_active = TRUE
		`
		err := app.DB.QueryRowContext(c, query, slug).Scan(
			&product.ID, &product.SKU, &product.Name, &product.Slug, &product.ImageURL, &product.ImageKey, &product.Description,
			&product.Quantity, &product.Price, &product.Taxable, &product.IsActive, &product.BrandID, &product.MerchantID, &product.Updated, &product.Created, &product.MerchantID)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{"message": "No product found."})
			} else {
				l.DebugF("Database query error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve product"})
			}
			return
		}

		query = `
		SELECT id, name, slug
		FROM categories
		JOIN product_categories ON categories.id = product_categories.category_id
		WHERE product_categories.product_id = $1;
		`
		rows, err := app.DB.QueryContext(c, query, product.ID)
		if err != nil {
			l.ErrorF("Error querying categories: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
			return
		}
		defer rows.Close()

		categories := []category.Category{}
		for rows.Next() {
			var cat category.Category
			err := rows.Scan(&cat.ID, &cat.Name, &cat.Slug)
			if err != nil {
				l.ErrorF("Error scanning category: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan categories"})
				return
			}
			categories = append(categories, cat)
		}
		var product_request = GetProduct{
			ID:          product.ID,
			SKU:         product.SKU,
			Name:        product.Name,
			Slug:        product.Slug,
			ImageURL:    []string{},
			Description: product.Description,
			Quantity:    product.Quantity,
			Price:       product.Price,
			Taxable:     product.Taxable,
			IsActive:    product.IsActive,
			Brand:       brand.Brand{ID: product.BrandID.UUID},
			Categories:  categories,
			MerchantID:  product.MerchantID,
			Created:     product.Created,
			Updated:     product.Updated.Time,
		}
		c.JSON(http.StatusOK, gin.H{"product": product_request})
	}
}

func SearchProductsByName(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		productName := c.Param("name")
		page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
		if err != nil || page < 1 {
			page = 1
		}
		limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
		if err != nil || limit < 1 {
			limit = 10
		}
		offset := (page - 1) * limit

		rows, err := app.DB.QueryContext(c, `
			SELECT id, sku, name, slug, image_url, description, quantity, price, merchant_id, created, updated
			FROM products
			WHERE name ILIKE $1 AND is_active = TRUE
			LIMIT $2 OFFSET $3
		`, "%"+productName+"%", limit, offset) // Case-insensitive search with ILIKE and wildcards

		if err != nil {
			l.ErrorF("Database query error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search products."}) // Generic error message for security
			return
		}

		defer rows.Close()

		var getProducts []GetProduct

		products := []Product{} // Initialize an empty slice to avoid null in the response
		for rows.Next() {
			var product Product
			err := rows.Scan(&product.ID, &product.SKU, &product.Name, &product.Slug, &product.ImageURL, &product.Description, &product.Quantity, &product.Price, &product.MerchantID, &product.Created, &product.Updated)

			if err != nil {
				l.ErrorF("Error scanning product: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan products data"}) // Generic error for security
				return
			}
			products = append(products, product) // Append to the slice
		}

		categories := make(map[uuid.UUID][]category.Category)
		productIds := []uuid.UUID{}
		for _, product := range products {
			productIds = append(productIds, product.ID)
		}

		if len(productIds) > 0 {
			rows, err := app.DB.QueryContext(c, `
				SELECT c.id, c.name, c.slug, pc.product_id as product_id
				FROM categories c
				JOIN product_categories pc ON c.id = pc.category_id
				WHERE pc.product_id = ANY($1)
			`, pq.Array(productIds)) // Use ANY() to match multiple values

			if err != nil {
				l.ErrorF("Error querying categories: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
				return
			}
			defer rows.Close()

			for rows.Next() {
				var cat category.Category
				var productID uuid.UUID
				err := rows.Scan(&cat.ID, &cat.Name, &cat.Slug, &productID)
				if err != nil {
					l.ErrorF("Error scanning category: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan categories"})
					return
				}
				categories[productID] = append(categories[productID], cat)

			}

		}

		for _, product := range products {
			getProduct := GetProduct{
				ID:          product.ID,
				SKU:         product.SKU,
				Name:        product.Name,
				Slug:        product.Slug,
				ImageURL:    []string{},
				Description: product.Description,
				Quantity:    product.Quantity,
				Price:       product.Price,
				Taxable:     product.Taxable,
				IsActive:    product.IsActive,
				Brand:       brand.Brand{ID: product.BrandID.UUID},
				Categories:  categories[product.ID],
				MerchantID:  product.MerchantID,
				Created:     product.Created,
				Updated:     product.Updated.Time,
			}
			getProducts = append(getProducts, getProduct)
		}

		c.JSON(http.StatusOK, gin.H{"products": getProducts, "page": page, "limit": limit}) // Return even if empty
	}
}

func FetchStoreProductsByFilters(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse query parameters and set default values if not provided.  Error handling for parsing is important.
		minPriceStr := c.DefaultQuery("min", "0")
		maxPriceStr := c.DefaultQuery("max", "0") // Or a very large number if no max limit
		ratingStr := c.DefaultQuery("rating", "0")
		categorySlug := c.Query("category")
		brandSlug := c.Query("brand")
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

		query := `
		SELECT p.id, p.sku, p.name, p.slug, p.image_url, p.description, p.quantity, p.price, p.taxable, p.is_active, p.brand_id, p.merchant_id, p.updated, p.created
		FROM products p
		LEFT JOIN product_categories pc ON p.id = pc.product_id
		WHERE p.is_active = true` // Base query

		args := []interface{}{}
		argIndex := 1

		if minPriceStr != "0" {
			minPrice, err := strconv.ParseFloat(minPriceStr, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'min' price"})
				return
			}
			query += fmt.Sprintf(" AND p.price >= $%d", argIndex)
			args = append(args, minPrice)
			argIndex++
		}

		if maxPriceStr != "0" {
			maxPrice, err := strconv.ParseFloat(maxPriceStr, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid max price"})
				return
			}
			query += fmt.Sprintf(" AND p.price <= $%d", argIndex)
			args = append(args, maxPrice)
			argIndex++
		}

		if ratingStr != "0" {
			rating, err := strconv.ParseFloat(ratingStr, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid minimum rating value"})
				return
			}
			query += fmt.Sprintf(` AND p.id IN (SELECT product_id FROM reviews GROUP BY product_id HAVING AVG(rating) >= $%d)`, argIndex)
			args = append(args, rating)
			argIndex++
		}

		if categorySlug != "" {
			query += fmt.Sprintf(" AND cat.slug = $%d", argIndex)
			args = append(args, categorySlug)
			argIndex++
		}

		if brandSlug != "" {
			query += fmt.Sprintf(" AND b.slug = $%d", argIndex)
			args = append(args, brandSlug)
			argIndex++
		}

		// Add sorting and pagination (ORDER BY, LIMIT, OFFSET)
		query += fmt.Sprintf(" ORDER BY p.created DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
		args = append(args, limit, (page-1)*limit)

		// Execute query
		// l.DebugF("%s, %v", query, args)
		rows, err := app.DB.QueryContext(c, query, args...)

		// ... Rest of the function is similar (error handling, rows.Close(), etc.)
		if err != nil {
			l.ErrorF("Error querying products: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
			return
		}

		defer rows.Close()

		products := make([]Product, 0) // Initialize as an empty slice
		productIds := []uuid.UUID{}

		for rows.Next() {
			var product Product

			err := rows.Scan(
				&product.ID, &product.SKU, &product.Name, &product.Slug, &product.ImageURL, &product.Description, &product.Quantity, &product.Price, &product.Taxable, &product.IsActive, &product.BrandID, &product.MerchantID, &product.Updated, &product.Created)

			if err != nil {
				l.DebugF("Error scanning products: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
				return
			}
			productIds = append(productIds, product.ID)
			products = append(products, product)
		}
		categories := make(map[uuid.UUID][]category.Category)

		if len(productIds) != 0 {
			query := `
			SELECT id, name, slug, product_id
			FROM categories
			JOIN product_categories ON categories.id = product_categories.category_id
			WHERE product_categories.product_id = ANY($1)
			`
			rows, err := app.DB.QueryContext(c, query, pq.Array(productIds))
			if err != nil {
				l.ErrorF("Error querying categories: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
				return
			}
			defer rows.Close()

			for rows.Next() {
				var cat category.Category
				var productID uuid.UUID
				err := rows.Scan(&cat.ID, &cat.Name, &cat.Slug, &productID)
				if err != nil {
					l.ErrorF("Error scanning category: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan categories"})
					return
				}
				l.DebugF("%#v", cat)
				categories[productID] = append(categories[productID], cat)
			}

		}

		var getProducts []GetProduct

		for _, product := range products {
			getProduct := GetProduct{
				ID:          product.ID,
				SKU:         product.SKU,
				Name:        product.Name,
				Slug:        product.Slug,
				ImageURL:    []string{},
				Description: product.Description,
				Quantity:    product.Quantity,
				Price:       product.Price,
				Taxable:     product.Taxable,
				IsActive:    product.IsActive,
				Brand:       brand.Brand{ID: product.BrandID.UUID},
				Categories:  categories[product.ID],
				MerchantID:  product.MerchantID,
				Created:     product.Created,
				Updated:     product.Updated.Time,
			}
			getProducts = append(getProducts, getProduct)
		}

		c.JSON(http.StatusOK, gin.H{"products": getProducts})

	}

}

func FetchProductNames(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := app.DB.QueryContext(c, "SELECT id, name FROM products") // Select only id and name
		if err != nil {
			l.DebugF("Error querying products: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product names"})
			return
		}
		defer rows.Close()

		products := []Product{}
		for rows.Next() {
			var product Product
			if err := rows.Scan(&product.ID, &product.Name); err != nil {
				l.DebugF("Error scanning product: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product names"})
				return
			}
			products = append(products, product)
		}

		c.JSON(http.StatusOK, gin.H{"products": products})
	}
}

func AddProduct(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input AddProductInput // Use the input struct
		// input.SKU = c.PostForm("sku")
		// input.Name = c.PostForm("name")
		// input.Slug = c.PostForm("slug")
		// input.Description = c.PostForm("description")
		// input.Quantity, _ = strconv.Atoi(c.PostForm("quantity"))
		// input.Price, _ = strconv.ParseFloat(c.PostForm("price"), 64)
		// input.Taxable, _ = strconv.ParseBool(c.PostForm("taxable"))
		// input.IsActive, _ = strconv.ParseBool(c.PostForm("is_active"))
		// input.BrandID, _ = uuid.Parse(c.PostForm("brand_id"))
		// input.CategoryID, _ = uuid.Parse(c.PostForm("category_id"))
		if err := c.ShouldBind(&input); err != nil {
			l.DebugF("Error binding input: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var imageUrl, imageKey string
		var err error
		file, _ := c.FormFile("image")
		if file != nil {
			if file.Size > 2<<20 { // 2MB
				c.JSON(http.StatusBadRequest, gin.H{"error": "Image size should be less than 2MB"})
				return
			}
			l.DebugF("%#v", file)
			imageUrl, imageKey, err = misc.S3Upload(file, app)
			if err != nil {
				l.ErrorF("Image upload failed: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Image upload failed"}) // More specific error message
				return
			}
		}

		var skuCount int
		err = app.DB.QueryRowContext(c, "SELECT COUNT(*) FROM products WHERE sku = $1", input.SKU).Scan(&skuCount)
		if err != nil {
			l.ErrorF("Failed to check SKU uniqueness: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check SKU"})
			return
		}
		if skuCount > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "This SKU is already in use."})
			return
		}

		// userIDStr := c.GetString("userID") // Assuming merchant ID is passed as user ID. Adjust this based on your authentication setup
		// userID, err := uuid.Parse(userIDStr)
		merchantIDStr := c.GetString("merchantID")
		merchantID, err := uuid.Parse(merchantIDStr)
		if err != nil {
			l.ErrorF("Invalid merchant ID: %v", err) // Log and return an error
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid merchant ID"})
			return
		}

		newProductID := uuid.New()

		// Build the dynamic query
		query := "INSERT INTO products (id, sku, name, slug, created, updated"
		values := "VALUES ($1, $2, $3, $4, $5, $6"
		args := []interface{}{newProductID, input.SKU, input.Name, input.Slug, time.Now(), time.Now()}
		argIndex := 7

		if imageUrl != "" {
			query += ", image_url, image_key"
			values += fmt.Sprintf(", $%d, $%d", argIndex, argIndex+1)
			args = append(args, imageUrl, imageKey)
			argIndex += 2
		}
		if input.Description != "" {
			query += ", description"
			values += fmt.Sprintf(", $%d", argIndex)
			args = append(args, input.Description)
			argIndex++
		}
		if input.Quantity != 0 {
			query += ", quantity"
			values += fmt.Sprintf(", $%d", argIndex)
			args = append(args, input.Quantity)
			argIndex++
		}
		if input.Price != 0 {
			query += ", price"
			values += fmt.Sprintf(", $%d", argIndex)
			args = append(args, input.Price)
			argIndex++
		}
		if input.Taxable {
			query += ", taxable"
			values += fmt.Sprintf(", $%d", argIndex)
			args = append(args, input.Taxable)
			argIndex++
		}
		if input.IsActive {
			query += ", is_active"
			values += fmt.Sprintf(", $%d", argIndex)
			args = append(args, input.IsActive)
			argIndex++
		}
		if input.BrandID != uuid.Nil {
			query += ", brand_id"
			values += fmt.Sprintf(", $%d", argIndex)
			args = append(args, input.BrandID)
			argIndex++
		}
		// TODO add category ID
		query += ", merchant_id) "
		values += fmt.Sprintf(", $%d)", argIndex)
		args = append(args, merchantID)

		query += values
		l.Debug(merchantIDStr)
		l.DebugF("%v", merchantID)
		_, err = app.DB.ExecContext(c, query, args...)
		if err != nil {
			l.ErrorF("Failed to insert product: %v", err) // Log the error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add product"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success":    true,
			"message":    "Product added successfully!",
			"product_id": newProductID, // Return the product ID
		})
	}
}

func FetchProducts(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {

		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})

			return

		}

		userRole := common.GetUserRole(role)

		userIDStr := c.GetString("userID")
		userID, err := uuid.Parse(userIDStr)

		if err != nil {

			l.DebugF("Error parsing merchant ID: %v", err) // Log the parsing error
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		var rows *sql.Rows
		if userRole == common.RoleMerchant {
			// Fetch products only for this merchant
			query := `
			SELECT id, sku, name, slug, image_url, image_key, description, quantity, price, taxable, is_active, brand_id, merchant_id, updated, created
			FROM products WHERE merchant_id = $1
			`
			rows, err = app.DB.QueryContext(c, query, userID)

		} else if userRole == common.RoleAdmin { // Add an admin case
			query := `
			SELECT id, sku, name, slug, image_url, image_key, description, quantity, price, taxable, is_active, brand_id, merchant_id, updated, created
			FROM products
			`

			rows, err = app.DB.QueryContext(c, query) // Fetch all product details

		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return

		}

		if err != nil {

			l.DebugF("Error querying products: %v", err) // Log the actual database error for better debugging
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
			return
		}

		defer rows.Close()
		products := make([]Product, 0) // Initialize as an empty slice

		for rows.Next() {
			var product Product

			err := rows.Scan(
				&product.ID, &product.SKU, &product.Name, &product.Slug, &product.ImageURL, &product.ImageKey, &product.Description, &product.Quantity, &product.Price, &product.Taxable, &product.IsActive, &product.BrandID, &product.MerchantID, &product.Updated, &product.Created)

			if err != nil {
				l.DebugF("Error scanning products: %v", err) // More specific error message
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
				return
			}

			products = append(products, product)
		}

		c.JSON(http.StatusOK, gin.H{"products": products})

	}
}

func FetchProduct(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {

		_userID, exists := c.MustGet("merchantID").(string) // Get merchant ID if available
		if !exists {
			_userID = "" // or handle the case where it's not available appropriately
		}

		merchantID, err := uuid.Parse(_userID) // Assuming merchant ID is a UUID.  If not, adjust this part.

		if err != nil && _userID != "" {
			l.DebugF("Invalid merchant ID: %s. Error: %v", _userID, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid merchant ID"})
			return
		}

		userRole, ok := c.MustGet("role").(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "failed to get user role"})
			return
		}
		role := common.GetUserRole(userRole)

		productIDStr := c.Param("id")
		productID, err := uuid.Parse(productIDStr)

		if err != nil {
			l.DebugF("Error parsing product ID: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
			return
		}

		var product Product

		if role == common.RoleMerchant {
			query := `
			SELECT id, sku, name, slug, image_url, image_key, description, quantity, price, taxable, is_active, brand_id, merchant_id, updated, created
			FROM products WHERE id = $1 AND merchant_id = $2
			`
			err = app.DB.QueryRowContext(c, query, productID, merchantID).Scan(&product.ID, &product.SKU, &product.Name, &product.Slug, &product.ImageURL, &product.ImageKey, &product.Description, &product.Quantity, &product.Price, &product.Taxable, &product.IsActive, &product.BrandID, &product.MerchantID, &product.Updated, &product.Created)
		} else if role == common.RoleAdmin {
			query := `
			SELECT id, sku, name, slug, image_url, image_key, description, quantity, price, taxable, is_active, brand_id, merchant_id, updated, created
			FROM products WHERE id = $1
			`

			err = app.DB.QueryRowContext(c, query, productID).Scan(&product.ID, &product.SKU, &product.Name, &product.Slug, &product.ImageURL, &product.ImageKey, &product.Description, &product.Quantity, &product.Price, &product.Taxable, &product.IsActive, &product.BrandID, &product.MerchantID, &product.Updated, &product.Created)

		}

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{"message": "Product not found"})
			} else {

				l.ErrorF("Failed to fetch product: %v", err) // Log error with more context
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product"})
			}

			return
		}

		c.JSON(http.StatusOK, gin.H{"product": product})
	}
}

func UpdateProduct(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// ... (get productID, userID, userRole - same as before)
		productIDStr := c.Param("id")
		productID, err := uuid.Parse(productIDStr)
		if err != nil {
			l.DebugF("Invalid product ID: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
			return
		}

		userRoleStr, ok := c.MustGet("role").(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "failed to get user role"})
			return
		}

		userRole := common.GetUserRole(userRoleStr)

		userIdStr, exists := c.MustGet("userID").(string)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		userID, err := uuid.Parse(userIdStr)
		if err != nil {
			l.DebugF("Invalid user ID: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		var updateProduct ProductUpdate
		if err := c.ShouldBindJSON(&updateProduct); err != nil {

			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"}) // Return a 400 error for invalid input
			return
		}

		ctx := context.Background()
		tx, err := app.DB.BeginTx(ctx, nil) // Start a transaction for data consistency.
		if err != nil {
			l.ErrorF("Error starting transaction: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})

			return

		}
		defer tx.Rollback() // Important: defer the rollback to handle potential errors

		updateQuery := "UPDATE products SET updated = $1"
		args := []interface{}{time.Now()}
		argIndex := 2

		if updateProduct.Name != nil {
			updateQuery += fmt.Sprintf(", name = $%d", argIndex)
			args = append(args, *updateProduct.Name)
			argIndex++
		}

		// ... similarly handle other fields (SKU, Slug, etc.) as needed
		if updateProduct.SKU != nil {
			updateQuery += fmt.Sprintf(", sku = $%d", argIndex)
			args = append(args, *updateProduct.SKU)
			argIndex++
		}

		if updateProduct.Description != nil {
			updateQuery += fmt.Sprintf(", description = $%d", argIndex)
			args = append(args, *updateProduct.Description)
			argIndex++
		}

		if updateProduct.Quantity != nil {
			updateQuery += fmt.Sprintf(", quantity = $%d", argIndex)
			args = append(args, *updateProduct.Quantity)
			argIndex++
		}

		if updateProduct.Price != nil {
			updateQuery += fmt.Sprintf(", price = $%d", argIndex)
			args = append(args, *updateProduct.Price)
			argIndex++
		}

		if updateProduct.Taxable != nil {
			updateQuery += fmt.Sprintf(", taxable = $%d", argIndex)
			args = append(args, *updateProduct.Taxable)
			argIndex++
		}

		if updateProduct.IsActive != nil {
			updateQuery += fmt.Sprintf(", is_active = $%d", argIndex)
			args = append(args, *updateProduct.IsActive)
			argIndex++
		}

		if updateProduct.BrandID != nil {
			updateQuery += fmt.Sprintf(", brand_id = $%d", argIndex)
			args = append(args, *updateProduct.BrandID)
			argIndex++
		}

		// TODO add category ID
		if updateProduct.Slug != nil {
			updateQuery += fmt.Sprintf(", slug = $%d", argIndex)
			args = append(args, *updateProduct.Slug)
			argIndex++
		}

		updateQuery += fmt.Sprintf(" WHERE id = $%d", argIndex)
		args = append(args, productID)

		if userRole == common.RoleMerchant {
			// Check if product belongs to merchant for authorization

			var productMerchantID uuid.UUID
			err = app.DB.QueryRowContext(c, "SELECT merchant_id FROM products WHERE id = $1", productID).Scan(&productMerchantID)
			if err != nil {
				if err == sql.ErrNoRows {

					c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
					return

				} else {

					l.DebugF("Error checking product ownership: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check product ownership"})
					return
				}

			}

			if productMerchantID != userID { // Authorization check
				c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to update this product"})
				return
			}

			updateQuery += fmt.Sprintf(" AND merchant_id = $%d", argIndex)
			args = append(args, userID)

		} else if userRole != common.RoleAdmin { // Admin can update any product
			c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"}) // Proper status code for authorization failure
			return
		}

		// Check for SKU and slug uniqueness only if being updated.
		if updateProduct.SKU != nil {
			var count int
			err = app.DB.QueryRowContext(c, "SELECT COUNT(*) FROM products WHERE sku = $1 AND id != $2", *updateProduct.SKU, productID).Scan(&count)
			if err != nil {
				l.ErrorF("Failed to check SKU uniqueness: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check SKU uniqueness."})
				return
			}
			if count > 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "SKU already exists."})
				return
			}
		}

		if updateProduct.Slug != nil {
			var slugCount int
			err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM products WHERE slug = $1 AND id != $2", *updateProduct.Slug, productID).Scan(&slugCount)
			if err != nil {
				l.ErrorF("Failed to check slug uniqueness: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check slug uniqueness."})
				return
			}
			if slugCount > 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Slug already exists."})
				return
			}
		}

		// Execute the update within the transaction
		_, err = tx.ExecContext(ctx, updateQuery, args...) // Use tx.ExecContext here
		if err != nil {
			l.DebugF("Error updating product: %v", err) // Log the specific error for better debugging.
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})

			return
		}

		if err = tx.Commit(); err != nil { // Commit only if no errors
			l.ErrorF("Failed to commit transaction: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Product updated successfully"})

	}
}

func UpdateProductStatus(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// ... (Get userID, userRole, productID)
		productIDStr := c.Param("id")
		productID, err := uuid.Parse(productIDStr)
		if err != nil {
			l.DebugF("Invalid product ID: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
			return
		}

		userRoleStr, ok := c.MustGet("role").(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "failed to get user role"})
			return
		}

		userRole := common.GetUserRole(userRoleStr)

		userIdStr, exists := c.MustGet("userID").(string)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		userID, err := uuid.Parse(userIdStr)
		if err != nil {
			l.DebugF("Invalid user ID: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		var updateData struct {
			IsActive bool `json:"is_active"`
		}
		if err := c.ShouldBindJSON(&updateData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})

			return
		}

		// Check if the user is a merchant or admin and authorize
		if userRole == common.RoleMerchant {
			var merchantID uuid.UUID
			err = app.DB.QueryRowContext(c, "SELECT merchant_id FROM products WHERE id = $1", productID).Scan(&merchantID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
				} else {
					l.DebugF("Error getting merchant ID: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify product"})
				}
				return
			}
			if merchantID != userID {
				c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to update this product"})
				return
			}
		} else if userRole != common.RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"})
			return
		}

		_, err = app.DB.ExecContext(c, `UPDATE products SET is_active = $1, updated = $2 WHERE id = $3`, updateData.IsActive, time.Now(), productID)

		if err != nil {

			l.ErrorF("Error updating product status: %v", err)                                        // Log error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product status"}) // Generic error message for security
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Product status updated successfully"})
	}
}

func DeleteProduct(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {

		// ... (Get productID, userID, userRole)
		productIDStr := c.Param("id")
		productID, err := uuid.Parse(productIDStr)
		if err != nil {
			l.DebugF("Invalid product ID: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
			return
		}

		userRoleStr, ok := c.MustGet("role").(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "failed to get user role"})
			return
		}

		userRole := common.GetUserRole(userRoleStr)

		userIdStr, exists := c.MustGet("userID").(string)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		userID, err := uuid.Parse(userIdStr)
		if err != nil {
			l.DebugF("Invalid user ID: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		ctx := context.Background()
		tx, err := app.DB.BeginTx(ctx, nil)

		if err != nil {

			l.DebugF("Failed to start trasaction: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
			return
		}
		defer tx.Rollback()

		if userRole == common.RoleMerchant {
			// Check if the product belongs to the merchant for authorization.

			var count int
			err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM products WHERE id = $1 AND merchant_id = $2", productID, userID).Scan(&count) // Check with transaction

			if err != nil {
				l.ErrorF("Error checking product: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify product."})
				return
			}
			if count == 0 { // No matching product found

				c.JSON(http.StatusNotFound, gin.H{"error": "Product not found or you are not authorized"}) // Correct handling for not found or unauthorized
				return
			}

		} else if userRole != common.RoleAdmin {

			c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized"}) // Correct status code
			return
		}

		res, err := tx.ExecContext(ctx, "DELETE FROM products WHERE id = $1", productID) // Corrected SQL statement to reflect database/sql usage
		if err != nil {

			l.DebugF("Error deleting product : %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product."}) // Generic message for security

			return
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {

			l.DebugF("Error getting rows affected : %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product."}) // Generic message for security

			return
		}

		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"message": "No product found."})
			return
		}

		if err = tx.Commit(); err != nil { // Commit transaction
			l.ErrorF("Transaction commit error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Product deleted successfully"})

	}
}
