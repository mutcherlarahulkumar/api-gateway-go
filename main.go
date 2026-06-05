package main

import (
	"api-gateway/middleware"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

var serviceMap = map[string]string{
	"user-service":    "http://localhost:8081",
	"product-service": "http://localhost:8082",
}

func main() {

	err := middleware.PingRedis()
	if err != nil {
		log.Fatal("Redis connection failed:", err)
	}

	router := gin.Default()
	// middleware
	router.Use(middleware.RateLimiterMiddleware)

	router.Any("/*path", GateWayHandler)

	router.Run(":8080")
}

func GateWayHandler(c *gin.Context) {

	parts := strings.Split(c.Request.RequestURI, "/")

	servicePrefix := parts[1]

	baseURL, exists := serviceMap[servicePrefix]
	if !exists {
		c.JSON(http.StatusNotFound, map[string]string{
			"message": "Invalid Route",
		})
		return
	}

	var finalParts []string

	for i := 2; i < len(parts); i++ {
		finalParts = append(finalParts, parts[i])
	}

	finalURL := strings.Join(finalParts, "/")

	c.Request.URL.Path = "/" + finalURL
	targetURL, _ := url.Parse(baseURL)
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.ServeHTTP(c.Writer, c.Request)

}
