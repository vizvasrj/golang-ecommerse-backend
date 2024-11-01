package brand

import (
	"src/pkg/conf"
	"src/pkg/middleware"
	"src/pkg/module/user"

	"github.com/gin-gonic/gin"
)

func SetupRouter(path string, r *gin.Engine, config *conf.Config) {
	brand_route := r.Group(path)

	{
		brand_route.POST("/add",
			middleware.AuthMiddleware(config),
			middleware.RoleCheck(user.RoleAdmin, user.RoleMerchant),
			AddBrand(config))

		brand_route.GET("/list", ListBrands(config))
		brand_route.GET("/", GetBrands(config))
		brand_route.GET("/:id", GetBrandByID(config))
		brand_route.GET("/list/select", ListSelectBrands(config))

		brand_route.PUT("/:id",
			middleware.AuthMiddleware(config),
			middleware.RoleCheck(user.RoleAdmin, user.RoleMerchant),
			UpdateBrand(config))

		brand_route.PUT("/:id/active",
			middleware.AuthMiddleware(config),
			middleware.RoleCheck(user.RoleAdmin, user.RoleMerchant),
			UpdateBrandActive(config))

		brand_route.DELETE("/delete/:id",
			middleware.AuthMiddleware(config),
			middleware.RoleCheck(user.RoleAdmin, user.RoleMerchant),
			DeleteBrand(config))
	}
}
