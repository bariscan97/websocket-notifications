package cacherepo

import (
	"context"
	"fmt"
	"strconv"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

type RedisClient struct {
	Rdb *redis.Client
}

type IRedisClient interface {
	CreateNotification(data map[string]interface{}) error
	GetNotifications(username string, page string) ([]map[string]interface{}, error)
	GetUnreadCount(username string) string
	IncUnreadCount(username string) error
	ResetUnreadCount(username string) error
	GetSubsByUsername(username string) ([]string, error)
}

func NewCacheClient(rdb *redis.Client) IRedisClient {
	return &RedisClient{
		Rdb: rdb,
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
		var notification map[string]interface{}
		if ok {
			for j := 0; j < len(doc); j += 2 {
				key, isKeyString := doc[j].(string)
				value, isValueString := doc[j+1].(string)
				if key == "username" {
					continue
				}
				if isKeyString && isValueString {
					notification[key] = value
				}
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
		return fmt.Errorf(err.Error())
	}
	num++
	redisCli.Rdb.Set(ctx, key, num, 0)
	return nil
}

func (redisCli *RedisClient) ResetUnreadCount(username string) error {
	ctx := context.Background()
	key := fmt.Sprintf("unread:%s", username)
	err := redisCli.Rdb.Set(ctx, key, 0, 0)
	return fmt.Errorf(err.Err().Error())
}

func (redisCli *RedisClient) GetSubsByUsername(username string) ([]string, error) {
	ctx := context.Background()
	key := fmt.Sprintf("subs:%s", username)
	return redisCli.Rdb.SMembers(ctx, key).Result()
}
