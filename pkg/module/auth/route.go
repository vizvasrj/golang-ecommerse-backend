package auth

import (
	"src/pkg/conf"
	"src/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(path string, r *gin.RouterGroup, config *conf.Config) {
	auth_route := r.Group(path)
	{
		auth_route.POST("/login", Login(config))
		auth_route.POST("/register", Register(config))
		auth_route.GET("/google", GoogleLogin(config))
		auth_route.POST("/google/callback", GoogleCallback(config))
		auth_route.POST("/forgot", ForgotPassword(config))
		auth_route.POST("/reset/:token", ResetPasswordFromToken(config))

		auth_route.POST("/reset",
			middleware.AuthMiddleware(config),
			ResetPassword(config),
		)
	}
}
