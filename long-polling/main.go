package main

import (
	"long-polling/longpolling_redis"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}
	longpolling_redis.StartLongPollingrRedisServer()
}
