package main

import (
	"api-gateway/constants"
	"api-gateway/middleware"
	"api-gateway/model"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var serviceMap = map[string]string{
	constants.UserService:    "http://localhost:8081",
	constants.ProductService: "http://localhost:8082",
}

var circuitBreakers = map[string]*model.CircuitBreaker{
	constants.UserService: {
		Mu:          sync.Mutex{},
		State:       model.Closed,
		Failures:    0,
		Threshold:   5,
		LastFailure: nil,
		CoolDown:    time.Minute, // 1 minute
	},
	constants.ProductService: {
		Mu:          sync.Mutex{},
		State:       model.Closed,
		Failures:    0,
		Threshold:   5,
		LastFailure: nil,
		CoolDown:    time.Minute, // 1 minute
	},
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

	cb := circuitBreakers[servicePrefix]

	cb.Mu.Lock() // lock before reading/writing
	if cb.State == model.Open {
		if time.Since(*cb.LastFailure) < cb.CoolDown {
			c.JSON(http.StatusServiceUnavailable, map[string]string{
				"message": "Service unavailable, try again later",
			})
			cb.Mu.Unlock()
			return
		} else {
			cb.State = model.HalfOpen
		}
	}
	cb.Mu.Unlock() // unlock when function exits

	// if the service prefix's circuit breaker has it "open" then return and do c.json
	// only if last failure and current duration is less than cool down
	// if grater than cool down then make it half-open

	var finalParts []string

	for i := 2; i < len(parts); i++ {
		finalParts = append(finalParts, parts[i])
	}

	finalURL := strings.Join(finalParts, "/")

	c.Request.URL.Path = "/" + finalURL
	targetURL, _ := url.Parse(baseURL)
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	proxy.ServeHTTP(c.Writer, c.Request)

	// if req is passing
	// if it is half-open -> close itt completely and set falure to 0

	// if req is failng
	/// if open change increment falure and update the last failure

	//  if half-open -> increase threshold

	statusCode := c.Writer.Status()
	cb.Mu.Lock()         // lock before reading/writing
	defer cb.Mu.Unlock() // unlock when function exits
	// failing req
	if statusCode/100 == 5 { // 5XX
		cb.Failures++
		now := time.Now()
		cb.LastFailure = &now

		if cb.State == model.HalfOpen {
			cb.State = model.Open
			cb.CoolDown = cb.CoolDown * 2
		} else if cb.Failures >= cb.Threshold {
			cb.State = model.Open
		}

	} else { // passing req
		if cb.State != model.Closed {
			cb.State = model.Closed
			cb.Failures = 0
			cb.LastFailure = nil
		}
	}

}
