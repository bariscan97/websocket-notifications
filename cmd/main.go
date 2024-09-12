package main

import (
	"fmt"
	"log"
	"notifications/cacherepo"
	"notifications/router"
	"notifications/worker"
	"notifications/ws"
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

	hub := ws.NewHub()

	wsHandler := ws.NewHandler(hub)

	router.InitRouter(wsHandler)

	cache := cacherepo.NewCacheClient(client)

	go worker.Worker(cache, hub)

	go hub.Run()

	router.Start(fmt.Sprintf("0.0.0.0:%s", os.Getenv("PORT")))
}
