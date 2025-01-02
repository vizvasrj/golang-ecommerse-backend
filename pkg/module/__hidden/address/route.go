package address

import (
	"src/pkg/conf"
	"src/pkg/middleware"
	address2 "src/pkg/module/address_2"

	"github.com/gin-gonic/gin"
)

func SetupRouter(path string, r *gin.RouterGroup, config *conf.Config) {
	address_route := r.Group(path)
	address_route.Use(middleware.AuthMiddleware(config))
	{
		address_route.POST("/add", address2.AddAddress(config))
		address_route.GET("", address2.GetAddresses(config))
		address_route.GET("/:id", address2.GetAddress(config))
		address_route.PUT("/:id", address2.UpdateAddress(config))
		address_route.DELETE("/delete/:id", address2.DeleteAddress(config))
		address_route.PUT("/default/:id", address2.SetDefaultAddress(config))

	}
}
