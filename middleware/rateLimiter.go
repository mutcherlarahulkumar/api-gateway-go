package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	RequestsPerMinute = 100
)

func RateLimiterMiddleware(c *gin.Context) {
	ctx := c.Request.Context()
	clientIP := c.ClientIP()

	count, err := rdb.Incr(ctx, "rate:"+clientIP).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Something went wrong!",
		})

		return
	}

	if count == 1 {
		rdb.Expire(ctx, "rate:"+clientIP, time.Minute)
	} else if count > RequestsPerMinute {
		c.JSON(http.StatusTooManyRequests, map[string]string{
			"message": "Rate limit exceeded!",
		})
		c.Abort()

		return
	}

	c.Next()

}
