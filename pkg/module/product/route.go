package product

import (
	"src/pkg/conf"
	"src/pkg/middleware"
	"src/pkg/module/user"

	"github.com/gin-gonic/gin"
)

func SetupRouter(path string, r *gin.Engine, app *conf.Config) {
	product_route := r.Group(path)
	{
		product_route.GET("/item:slug", GetProductBySlug(app))
		product_route.GET("/list/search/:name", SearchProductsByName(app))
		product_route.GET("/list", FetchStoreProductsByFilters(app))
		product_route.GET("/list/select", FetchProductNames(app))

		product_route.POST("/add",
			middleware.AuthMiddleware(app),
			middleware.RoleCheck(user.RoleMerchant, user.RoleAdmin),
			AddProduct(app))

		product_route.GET("/",
			middleware.AuthMiddleware(app),
			middleware.RoleCheck(user.RoleMerchant, user.RoleAdmin),
			FetchProducts(app))

		product_route.PUT(":/id",
			middleware.AuthMiddleware(app),
			middleware.RoleCheck(user.RoleMerchant, user.RoleAdmin),
			UpdateProduct(app))

		product_route.PUT("/:id/active",
			middleware.AuthMiddleware(app),
			middleware.RoleCheck(user.RoleMerchant, user.RoleAdmin),
			UpdateProductStatus(app))

		product_route.DELETE("/delete/:id",
			middleware.AuthMiddleware(app),
			middleware.RoleCheck(user.RoleMerchant, user.RoleAdmin),
			DeleteProduct(app))
	}

}
