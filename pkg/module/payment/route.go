package payment

import (
	"src/pkg/conf"

	"github.com/gin-gonic/gin"
)

func SetupRouter(path string, r *gin.RouterGroup, app *conf.Config) {
	paymentRoute := r.Group(path)
	{
		paymentRoute.POST("/webhook", handleRazorPayWebhook(app))
	}

}
