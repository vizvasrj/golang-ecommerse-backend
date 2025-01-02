package user

import (
	"src/common"
	"src/pkg/conf"
	"src/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(path string, r *gin.RouterGroup, app *conf.Config) {
	userRoute := r.Group(path)
	{
		userRoute.GET("/search",
			middleware.AuthMiddleware(app),
			middleware.RoleCheck(common.RoleAdmin, common.RoleMerchant),
			SearchUsers(app))
		userRoute.GET("",
			middleware.AuthMiddleware(app),
			FetchUsers(app))

		userRoute.GET("/me",
			middleware.AuthMiddleware(app),
			GetCurrentUser(app))

		userRoute.PUT("",
			middleware.AuthMiddleware(app),
			UpdateUserProfile(app))

	}
}
