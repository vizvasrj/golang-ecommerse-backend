package review

import (
	"src/common"
	"src/pkg/conf"
	"src/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter(path string, r *gin.RouterGroup, config *conf.Config) {
	reviewRoute := r.Group(path)
	{
		reviewRoute.POST("/add",
			middleware.AuthMiddleware(config),
			AddReview(config))

		reviewRoute.GET("", GetAllReviews(config))
		reviewRoute.GET("/:slug", GetProductReviewsBySlug(config))

		reviewRoute.PUT("/:id",
			middleware.AuthMiddleware(config),
			UpdateReview(config))

		reviewRoute.PUT("/approve/:reviewId",
			middleware.AuthMiddleware(config),
			middleware.RoleCheck(common.RoleAdmin, common.RoleMerchant),
			ApproveReview(config))

		reviewRoute.PUT("/reject/:reviewId",
			middleware.AuthMiddleware(config),
			middleware.RoleCheck(common.RoleAdmin, common.RoleMerchant),
			ApproveReview(config))

		reviewRoute.DELETE("/delete/:id",
			middleware.AuthMiddleware(config),
			DeleteReview(config))

	}
}
