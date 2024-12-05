package payment

// import (
// 	"net/http"

// 	"github.com/go-redis/redis"
// 	"github.com/gin-gonic/gin"
// )

// func DuplicateEventMiddleware(redisClient *redis.Client) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		eventID := c.Request.Header.Get("x-razorpay-event-id")
// 		if eventID != "" {
// 			// Check if the event ID exists in Redis
// 			exists, err := redisClient.Exists(eventID).Result()
// 			if err != nil {
// 				// Handle Redis error
// 				c.AbortWithStatus(http.StatusInternalServerError)
// 				return
// 			}
// 			if exists == 1 {
// 				// Event ID already exists in Redis, so it's a duplicate
// 				c.AbortWithStatus(http.StatusConflict)
// 				return
// 			}
// 		}

// 		// Event ID doesn't exist in Redis, continue processing
// create a new key in Redis with the event ID
// 		c.Next()
// 	}
// }
