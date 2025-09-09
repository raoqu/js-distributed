package util

import (
	"context"
	"log"

	"main/config"

	"github.com/redis/go-redis/v9"
)

const DEFAULT_HSET_GROUP = "gotask"

type RedisClient struct {
	Client *redis.Client
}

var (
	ctx         = context.Background()
	RedisData   *RedisClient
	RedisConfig *RedisClient
)

func CreateRedisConn(db int) (*redis.Client, error) {
	// Initialize Redis client
	addr := config.CONFIG.Database.Redis.Addr

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       db, // use default DB
	})

	// Test Redis connection
	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Printf("Warning: Could not connect to Redis: %v", err)
		log.Printf("Redis operations will fail. Please ensure Redis is running on localhost:6379")
		return nil, err
	} else {
		log.Printf("Successfully connected to Redis %s, db %d\n", addr, db)
	}
	return client, nil
}

func InitRedisClient() error {
	client, err := CreateRedisConn(config.CONFIG.Database.Redis.DBConfig)
	if err != nil {
		return err
	} else {
		RedisConfig = &RedisClient{Client: client}
	}
	client, err = CreateRedisConn(config.CONFIG.Database.Redis.DB)
	if err != nil {
		return err
	} else {
		RedisData = &RedisClient{Client: client}
	}
	return nil
}

func (rc *RedisClient) Set(key string, value string) error {
	return rc.Client.Set(ctx, key, value, 0).Err()
}

func (rc *RedisClient) Get(key string) (string, error) {
	return rc.Client.Get(ctx, key).Result()
}

func (rc *RedisClient) SetHValue(group string, key string, value string) error {
	group = getGroupName(group)

	// HSET group key value
	return rc.Client.HSet(ctx, group, key, value).Err()
}

func (rc *RedisClient) GetHValue(group string, key string) (string, error) {
	group = getGroupName(group)

	// HGET group key
	value, err := rc.Client.HGet(ctx, group, key).Result()
	if err != nil {
		return "", err
	}

	return value, nil
}

func (rc *RedisClient) HKeys(group string) ([]string, error) {
	group = getGroupName(group)

	keys, err := rc.Client.HKeys(ctx, group).Result()
	if err != nil {
		return nil, err
	}

	return keys, nil
}

func (rc *RedisClient) HDel(group string, key string) error {
	group = getGroupName(group)

	return rc.Client.HDel(ctx, group, key).Err()
}

func getGroupName(group string) string {
	if group == "" {
		return DEFAULT_HSET_GROUP
	}
	return group
}

// SAdd adds one or more members to a set stored at key
func (rc *RedisClient) SAdd(key string, members ...interface{}) (int64, error) {
	result, err := rc.Client.SAdd(ctx, key, members...).Result()
	if err != nil {
		return 0, err
	}
	return result, nil
}

// SRem removes one or more members from a set stored at key
func (rc *RedisClient) SRem(key string, members ...interface{}) (int64, error) {
	result, err := rc.Client.SRem(ctx, key, members...).Result()
	if err != nil {
		return 0, err
	}
	return result, nil
}

// SCard returns the number of elements in the set stored at key
func (rc *RedisClient) SCard(key string) (int64, error) {
	result, err := rc.Client.SCard(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return result, nil
}

// SMembers returns all the members of the set value stored at key
func (rc *RedisClient) SMembers(key string) ([]string, error) {
	result, err := rc.Client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return result, nil
}
