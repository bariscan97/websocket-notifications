package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/rabbitmq/amqp091-go"
)

func Worker(cache *RedisClient, hub *Hub) {

	connString := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		os.Getenv("RABBITMQ_USER"),
		os.Getenv("RABBITMQ_PASSWORD"),
		os.Getenv("RABBITMQ_HOST"),
		os.Getenv("RABBITMQ_PORT"),
	)

	conn, err := amqp091.Dial(connString)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()
	err = ch.ExchangeDeclare(
		"notifications",
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare an exchange: %v", err)
	}
	q, err := ch.QueueDeclare(
		"websocket",
		false,
		false,
		true,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}
	err = ch.QueueBind(
		q.Name,
		"",
		"notifications",
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to bind a queue: %v", err)
	}
	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	go cache.NotifyQueue(hub)

	for msg := range msgs {

		var queueMessage QueueMessage
		err := json.Unmarshal(msg.Body, &queueMessage)
		if err != nil {
			log.Printf("Failed to unmarshal JSON: %v", err)
			continue
		}

		cache.Queue <- &queueMessage

		msg.Ack(true)

		fmt.Printf("Received a message: %+v\n", queueMessage)
	}

}
