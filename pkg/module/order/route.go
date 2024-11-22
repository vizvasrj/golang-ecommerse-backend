package order

import (
	"src/common"
	"src/pkg/conf"
	"src/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoute(path string, r *gin.RouterGroup, app *conf.Config) {

	order_route := r.Group(path)
	{
		order_route.POST("/add",
			middleware.AuthMiddleware(app),
			AddOrder(app))

		order_route.GET("/search",
			middleware.AuthMiddleware(app),
			SearchOrders(app))

		order_route.GET("",
			middleware.AuthMiddleware(app),
			FetchOrders(app))

		order_route.GET("/me",
			middleware.AuthMiddleware(app),
			FetchUserOrders(app))

		order_route.GET("/:orderId",
			middleware.AuthMiddleware(app),
			FetchOrder(app))

		order_route.DELETE("/cancel/:orderId",
			middleware.AuthMiddleware(app),
			CancelOrder(app))

		order_route.PUT("/status/item/:itemId",
			middleware.AuthMiddleware(app),
			middleware.RoleCheck(common.RoleAdmin, common.RoleMerchant),
			UpdateItemStatus(app))

	}
}
