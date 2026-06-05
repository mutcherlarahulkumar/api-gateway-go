package middleware

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var rdb = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "",
	DB:       0,
})

func PingRedis() error {
	return rdb.Ping(context.Background()).Err()
}
