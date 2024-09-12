package worker

import (
	"encoding/json"
	"fmt"
	"log"
	"notifications/cacherepo"
	"notifications/ws"
	"os"
	"github.com/rabbitmq/amqp091-go"
)

type QueueMessage struct {
	From    string `json:"from"`
	Content string `json:"content"`
}

func Worker(cache cacherepo.IRedisClient, hub *ws.Hub) {

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
		"notifications",   // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare an exchange: %v", err)
	}
	q, err := ch.QueueDeclare(
		"websocket",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}
	err = ch.QueueBind(
		q.Name, // queue name
		"",     // routing key
		"notifications", // exchange
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

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			var queueMessage QueueMessage
			err := json.Unmarshal(d.Body, &queueMessage)
			if err != nil {
				log.Printf("Failed to unmarshal JSON: %v", err)
				continue
			}
			var subscribers []string
			subscribers, err = cache.GetSubsByUsername(queueMessage.From)
			if err != nil {
				log.Printf("Error: %v", err.Error())
			}
			for _, sub := range subscribers {
				go func(subscriber string) {
					cache.CreateNotification(map[string]interface{}{
						"username": subscriber,
						"from":     queueMessage.From,
						"content":  queueMessage.Content,
					})
					hub.Broadcast <- &ws.Message{
						UnreadCount: cache.GetUnreadCount(subscriber),
						Username:    subscriber,
					}
				}(sub)
			}
			fmt.Printf("Received a message: %+v\n", queueMessage)
		}
	}()

	fmt.Println("Waiting for messages. To exit press CTRL+C")
	<-forever

}
