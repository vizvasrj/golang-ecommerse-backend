package brand

import (
	"src/common"
	"src/pkg/conf"
	"src/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(path string, r *gin.RouterGroup, config *conf.Config) {
	brand_route := r.Group(path)

	{
		brand_route.POST("/add",
			middleware.AuthMiddleware(config),
			middleware.RoleCheck(common.RoleAdmin, common.RoleMerchant),
			AddBrand(config))

		brand_route.GET("/list", ListBrands(config))
		brand_route.GET("", ListBrands(config))
		brand_route.GET("/:id", GetBrandByID(config))
		brand_route.GET("/list/select", ListSelectBrands(config))

		brand_route.PUT("/:id",
			middleware.AuthMiddleware(config),
			middleware.RoleCheck(common.RoleAdmin, common.RoleMerchant),
			UpdateBrand(config))

		brand_route.PUT("/:id/active",
			middleware.AuthMiddleware(config),
			middleware.RoleCheck(common.RoleAdmin, common.RoleMerchant),
			UpdateBrandActive(config))

		brand_route.DELETE("/delete/:id",
			middleware.AuthMiddleware(config),
			middleware.RoleCheck(common.RoleAdmin, common.RoleMerchant),
			DeleteBrand(config))
	}
}
