package user

import (
	"src/common"
	"src/pkg/conf"
	"src/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter_Merchant(path string, r *gin.RouterGroup, app *conf.Config) {
	merchant := r.Group(path)
	{
		merchant.POST("/add", AddMerchant(app))
		merchant.GET("/search",
			middleware.AuthMiddleware(app),
			middleware.RoleCheck(common.RoleAdmin),
			SearchMerchants(app))

		merchant.GET("",
			middleware.AuthMiddleware(app),
			middleware.RoleCheck(common.RoleAdmin),
			FetchAllMerchants(app))

		merchant.PUT("/:id/active",
			middleware.AuthMiddleware(app),
			middleware.RoleCheck(common.RoleAdmin, common.RoleMerchant),
			DisableMerchantAccount(app))

		// merchant.PUT("/approve/:id",
		// 	middleware.AuthMiddleware(app),
		// 	middleware.RoleCheck(RoleAdmin),
		// 	ApproveMerchant(app))

	}
}
