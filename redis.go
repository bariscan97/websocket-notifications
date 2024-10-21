package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type RedisClient struct {
	Rdb   *redis.Client
	Queue chan *QueueMessage
}

type IRedisClient interface {
	CreateNotification(data map[string]interface{}) error
	GetNotifications(username string, page string) ([]map[string]interface{}, error)
	GetUnreadCount(username string) string
	IncUnreadCount(username string) error
	ResetUnreadCount(username string) error
	GetSubsByUsername(username string) ([]string, error)
	NotifyQueue(queueMessage <-chan QueueMessage)
}

func NewCacheClient(rdb *redis.Client) *RedisClient {
	return &RedisClient{
		Rdb:   rdb,
		Queue: make(chan *QueueMessage, 10),
	}
}

func (redisCli *RedisClient) CreateNotification(data map[string]interface{}) error {
	ctx := context.Background()
	newUUID := uuid.New()
	key := fmt.Sprintf("notification:%s", newUUID.String())
	redisCli.IncUnreadCount(data["username"].(string))
	err := redisCli.Rdb.HSet(ctx, key, data)
	return err.Err()
}

func (redisCli *RedisClient) GetNotifications(username string, page string) ([]map[string]interface{}, error) {
	ctx := context.Background()
	query := fmt.Sprintf("@username:%s", username)
	var start int

	num, _ := strconv.Atoi(page)

	start = num
	pageSize := 15

	result, _ := redisCli.Rdb.Do(ctx, "FT.SEARCH", "idx:notifications", query, "SORTBY", "created_at", "DESC", "LIMIT", strconv.Itoa(start*pageSize), strconv.Itoa(pageSize)).Result()

	resultArray, _ := result.([]interface{})

	var notifications []map[string]interface{}

	for i := 1; i < len(resultArray); i++ {
		doc, ok := resultArray[i].([]interface{})
		notification := make(map[string]interface{})
		if ok {
			for j := 0; j < len(doc); j += 2 {
				key, _ := doc[j].(string)
				value, _ := doc[j+1].(interface{})
				notification[key] = value
			}
			notifications = append(notifications, notification)
		}
	}
	return notifications, nil
}

func (redisCli *RedisClient) GetUnreadCount(username string) string {
	ctx := context.Background()
	key := fmt.Sprintf("unread:%s", username)
	result, _ := redisCli.Rdb.Get(ctx, key).Result()
	return result
}

func (redisCli *RedisClient) IncUnreadCount(username string) error {
	ctx := context.Background()
	key := fmt.Sprintf("unread:%s", username)
	count := redisCli.GetUnreadCount(username)
	num, err := strconv.Atoi(count)
	if err != nil {
		return err
	}
	num++
	redisCli.Rdb.Set(ctx, key, num, 0)
	return nil
}

func (redisCli *RedisClient) ResetUnreadCount(username string) error {
	ctx := context.Background()
	key := fmt.Sprintf("unread:%s", username)
	err := redisCli.Rdb.Set(ctx, key, 0, 0)
	return err.Err()
}

func (redisCli *RedisClient) GetSubsByUsername(username string) ([]string, error) {
	ctx := context.Background()
	key := fmt.Sprintf("subs:%s", username)
	return redisCli.Rdb.SMembers(ctx, key).Result()
}

func (redisCli *RedisClient) NotifyQueue(hub *Hub) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		os.Exit(1)
	}()
loop:
	for {
		select {

		case <-ticker.C:
			if err := redisCli.Rdb.Ping(context.Background()).Err(); err != nil {
				log.Printf("Error:  %v", err)
				break loop
			}

		case msg, ok := <-redisCli.Queue:
			if !ok {
				log.Println("QueueMessage channel closed")
				break loop
			}
			var subscribers []string
			subscribers, err := redisCli.GetSubsByUsername(msg.From)
			if err != nil {
				log.Printf("Error while getting subscribers %v", err)
				continue
			}
			for _, subscriber := range subscribers {
				redisCli.CreateNotification(map[string]interface{}{
					"username": subscriber,
					"from":     msg.From,
					"content":  msg.Content,
				})

				hub.Emitter <- &Emit{
					User_slug: subscriber,
				}
			}

		}
	}

}
