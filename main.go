package main

import (
	"fmt"
	"log"
	"os"
    "github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	
	addr := fmt.Sprintf(
		"%s:%s",
		os.Getenv("REDIS_HOST"),
		os.Getenv("REDIS_PORT"),
	)
	
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	cache := NewCacheClient(client)

	hub := NewHub(cache)

    wsHandler := NewHandler(hub)

	InitRouter(wsHandler)

	go Worker(cache, hub)

	go hub.Run()

	Start(fmt.Sprintf("0.0.0.0:%s", os.Getenv("PORT")))
}
