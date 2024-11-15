package product

import (
	"src/common"
	"src/pkg/conf"
	"src/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(path string, r *gin.RouterGroup, app *conf.Config) {
	product_route := r.Group(path)
	{
		product_route.GET("/item/:slug", GetProductBySlug(app))
		product_route.GET("/list/search/:name", SearchProductsByName(app))
		product_route.GET("/list", FetchStoreProductsByFilters(app))
		product_route.GET("/list/select", FetchProductNames(app))

		product_route.POST("/add",
			middleware.AuthMiddleware(app),
			middleware.RoleCheck(common.RoleMerchant, common.RoleAdmin),
			AddProduct(app))

		product_route.GET("",
			middleware.AuthMiddleware(app),
			middleware.RoleCheck(common.RoleMerchant, common.RoleAdmin),
			FetchProducts(app))

		product_route.GET("/:id",
			middleware.AuthMiddleware(app),
			middleware.RoleCheck(common.RoleMerchant, common.RoleAdmin),
			FetchProduct(app))

		product_route.PUT("/:id",
			middleware.AuthMiddleware(app),
			middleware.RoleCheck(common.RoleMerchant, common.RoleAdmin),
			UpdateProduct(app))

		product_route.PUT("/:id/active",
			middleware.AuthMiddleware(app),
			middleware.RoleCheck(common.RoleMerchant, common.RoleAdmin),
			UpdateProductStatus(app))

		product_route.DELETE("/delete/:id",
			middleware.AuthMiddleware(app),
			middleware.RoleCheck(common.RoleMerchant, common.RoleAdmin),
			DeleteProduct(app))
	}

}
