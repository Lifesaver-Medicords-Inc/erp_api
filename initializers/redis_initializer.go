package initializers

import "github.com/redis/go-redis/v9"

var RC *redis.Client

func InitRedis() {
	RC = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
}
