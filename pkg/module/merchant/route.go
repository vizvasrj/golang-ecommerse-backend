package merchant

import (
	"src/pkg/conf"
	"src/pkg/middleware"
	"src/pkg/module/user"

	"github.com/gin-gonic/gin"
)

func SetupRouter(path string, r *gin.Engine, app *conf.Config) {
	merchant := r.Group(path)
	{
		merchant.POST("/add", AddMerchant(app))
		merchant.GET("/search",
			middleware.AuthMiddleware(app),
			middleware.RoleCheck(user.RoleAdmin),
			SearchMerchants(app))

		merchant.GET("/",
			middleware.AuthMiddleware(app),
			middleware.RoleCheck(user.RoleAdmin),
			FetchAllMerchants(app))

		merchant.PUT("/:id/active",
			middleware.AuthMiddleware(app),
			middleware.RoleCheck(user.RoleAdmin, user.RoleMerchant),
			DisableMerchantAccount(app))

		merchant.PUT("/approve/:id",
			middleware.AuthMiddleware(app),
			middleware.RoleCheck(user.RoleAdmin),
			ApproveMerchant(app))

	}
}
