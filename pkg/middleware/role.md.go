package middleware

import (
	"net/http"
	"src/common"

	"github.com/gin-gonic/gin"
)

func RoleCheck(allowedRoles ...common.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Role not found"})
			c.Abort()
			return
		}
		// l.DebugF("Role check: %#v", common.GetUserRole(role))
		userRole := common.GetUserRole(role)
		// if !ok {
		// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid role type"})
		// 	c.Abort()
		// 	return
		// }

		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		c.Abort()
	}
}
