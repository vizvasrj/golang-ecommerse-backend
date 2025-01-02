package category

import (
	"src/common"
	"src/pkg/conf"
	"src/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoute(path string, r *gin.RouterGroup, app *conf.Config) {
	category_route := r.Group(path)
	{
		category_route.POST("/add",
			middleware.AuthMiddleware(app),
			middleware.RoleCheck(common.RoleAdmin),
			AddCategory(app))

		category_route.GET("/list",
			ListCategories(app))

		category_route.GET("", FetchCategories(app))

		category_route.GET("/:id", FetchCategory(app))

		category_route.PUT("/:id",
			middleware.AuthMiddleware(app),
			middleware.RoleCheck(common.RoleAdmin),
			UpdateCategory(app))

		category_route.PUT("/:id/active",
			middleware.AuthMiddleware(app),
			middleware.RoleCheck(common.RoleAdmin),
			UpdateCategoryStatus(app))

		category_route.DELETE("/delete/:id",
			middleware.AuthMiddleware(app),
			middleware.RoleCheck(common.RoleAdmin),
			DeleteCategory(app))

		category_route.PUT("/product/:product_id/add",
			middleware.AuthMiddleware(app),
			middleware.RoleCheck(common.RoleAdmin),
			AddProductToCategory(app))

	}
}
