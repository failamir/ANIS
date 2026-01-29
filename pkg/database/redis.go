package database

import (
	"context"
	"fmt"
	"log"

	"auth-service/internal/config"

	"github.com/redis/go-redis/v9"
)

var Rdb *redis.Client
var Ctx = context.Background()

func ConnectRedis(cfg *config.Config) {
	Rdb = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		DB:   0, // usage of default DB
	})

	_, err := Rdb.Ping(Ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Connected to Redis")
}
