package cart2

import (
	"src/pkg/conf"
	"src/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoute(path string, r *gin.RouterGroup, app *conf.Config) {
	cart_route := r.Group(path)
	{
		cart_route.POST("/add",
			middleware.AuthMiddleware(app),
			AddToCart(app))

		cart_route.DELETE("/delete/:cartId",
			middleware.AuthMiddleware(app),
			DeleteCart(app))

		cart_route.POST("/add/:cartId",
			middleware.AuthMiddleware(app),
			AddProductToCart(app))

		cart_route.POST("/add_or_update",
			middleware.AuthOrNotMiddleware(app),
			AddProductToCartV2(app))

		cart_route.DELETE("/delete/:cartId/:productId",
			middleware.AuthMiddleware(app),
			RemoveProductFromCart(app))

		cart_route.GET("/:cartId",
			middleware.AuthOrNotMiddleware(app),
			GetCartByCartID(app))

	}
}
