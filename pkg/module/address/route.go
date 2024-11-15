package address

import (
	"src/pkg/conf"
	"src/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(path string, r *gin.RouterGroup, config *conf.Config) {
	address_route := r.Group(path)
	address_route.Use(middleware.AuthMiddleware(config))
	{
		address_route.POST("/add", AddAddress(config))
		address_route.GET("", GetAddresses(config))
		address_route.GET("/:id", GetAddress(config))
		address_route.PUT("/:id", UpdateAddress(config))
		address_route.DELETE("/delete/:id", DeleteAddress(config))
		address_route.PUT("/default/:id", SetDefaultAddress(config))

	}
}
