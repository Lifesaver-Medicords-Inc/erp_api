package initializers

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

var RC *redis.Client

func InitRedis() {
	RC = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	// pings redis if it is running, if not throw an error and exit the app.
	_, err := RC.Conn().Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}
}
