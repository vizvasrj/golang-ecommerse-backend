package user

import (
	"src/pkg/conf"
	"src/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(path string, app *conf.Config, r *gin.Engine) {
	userRoute := r.Group(path)
	{
		userRoute.GET("/search",
			middleware.AuthMiddleware(app),
			middleware.RoleCheck(RoleAdmin, RoleMerchant),
			SearchUsers(app))
		userRoute.GET("/",
			middleware.AuthMiddleware(app),
			FetchUsers(app))

		userRoute.GET("/me",
			middleware.AuthMiddleware(app),
			GetCurrentUser(app))

		userRoute.PUT("/",
			middleware.AuthMiddleware(app),
			UpdateUserProfile(app))

	}
}
