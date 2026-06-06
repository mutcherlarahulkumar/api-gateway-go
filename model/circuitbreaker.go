package model

import (
	"sync"
	"time"
)

type ServiceStatus string

const (
	// 1. closed
	// 2. open
	// 3. half-open
	Closed   ServiceStatus = "closed"
	Open     ServiceStatus = "open"
	HalfOpen ServiceStatus = "half-open"
)

type CircuitBreaker struct {
	Mu          sync.Mutex
	State       ServiceStatus // "closed", "open", "half-open"
	Failures    int           // current failure count
	Threshold   int           // max failures before opening (e.g. 5)
	LastFailure *time.Time    // when did it last fail
	CoolDown    time.Duration // how long to wait before half-open (e.g. 30s)
}
