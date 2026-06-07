package main

import (
	"api-gateway/handlers"
	"api-gateway/middleware"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {

	err := middleware.PingRedis()
	if err != nil {
		log.Fatal("Redis connection failed:", err)
	}

	router := gin.Default()
	// middleware
	router.Use(middleware.RateLimiterMiddleware)

	router.Any("/*path", handlers.GateWayHandler)

	router.Run(":8080")
}
