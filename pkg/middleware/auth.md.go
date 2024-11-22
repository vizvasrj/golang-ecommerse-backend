package middleware

import (
	"net/http"
	"src/pkg/conf"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(app *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := ValidateToken(app, tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// l.DebugF("claims %#v", claims)

		c.Set("userID", claims.Uid)
		c.Set("role", claims.Role)
		c.Set("email", claims.Email)
		c.Set("firstname", claims.FirstName)
		c.Set("lastname", claims.LastName)
		c.Set("merchantID", claims.MerchantID)

		c.Next()
	}
}
