package model

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"sync"
	"time"
)

var (
	redisClient *redis.Client
	redisOnce   sync.Once
)

type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

func InitRedisSentinel(sentinelAddr []string, passwd, name string, db int) *redis.Client {
	if redisClient == nil {
		redisOnce.Do(func() {
			redisClient = redis.NewFailoverClient(&redis.FailoverOptions{
				MasterName:    name,
				SentinelAddrs: sentinelAddr,
				Password:      passwd, // no password set
				DB:            db,     // use default DB
				DialTimeout:   10 * time.Second,
				ReadTimeout:   30 * time.Second,
				WriteTimeout:  30 * time.Second,
				PoolSize:      500,
				PoolTimeout:   30 * time.Second,
				MaxRetries:    2,
				IdleTimeout:   5 * time.Minute,
			})
			pong, err := redisClient.Ping(context.Background()).Result()
			if err != nil {
				fmt.Println("InitRedisSentinel failed: ", err.Error())
			}
			fmt.Println(pong, err)
		})
	}
	return redisClient
}

func InitRedis(addr string, passwd string, db int) *redis.Client {
	if redisClient == nil {
		redisOnce.Do(func() {
			redisClient = redis.NewClient(&redis.Options{
				Addr:         addr,
				Password:     passwd, // no password set
				DB:           db,     // use default DB
				DialTimeout:  10 * time.Second,
				ReadTimeout:  30 * time.Second,
				WriteTimeout: 30 * time.Second,
				PoolSize:     500,
				PoolTimeout:  30 * time.Second,
				MaxRetries:   2,
				IdleTimeout:  5 * time.Minute,
			})
			pong, err := redisClient.Ping(context.Background()).Result()
			if err != nil {
				fmt.Println("InitRedis failed: ", err.Error())
			}
			fmt.Println(pong, err)
		})
	}
	return redisClient
}

func NewRedisClient() (*RedisClient, error) {
	return &RedisClient{
		client: redisClient,
		ctx:    context.Background(),
	}, nil
}

func (r *RedisClient) GetConn() *redis.Client {
	return r.client
}

func (r *RedisClient) Set(k string, data interface{}) error {
	err := r.client.Set(r.ctx, k, data, 0).Err()
	if err != nil {
		fmt.Println("Failed to set data in Redis: ", err)
		return err
	}
	return nil
}

func (r *RedisClient) Get(k string) (interface{}, error) {
	val, err := r.client.Get(r.ctx, k).Result()
	if err != nil {
		fmt.Println("Failed to get data from Redis: ", err)
		return nil, err
	}
	return val, nil
}

func (r *RedisClient) HSet(k string, data map[string]interface{}) error {
	err := r.client.HSet(r.ctx, k, data).Err()
	if err != nil {
		fmt.Println("Failed to set data in Redis: ", err)
		return err
	}
	return nil
}

func (r *RedisClient) HSetBy(k string, field string, value interface{}) error {
	err := r.client.HSet(r.ctx, k, field, value).Err()
	if err != nil {
		fmt.Println("Failed to set data in Redis: ", err)
		return err
	}
	return nil
}

func (r *RedisClient) HIncrBy(k string, field string, value int64) error {
	_, err := r.client.HIncrBy(r.ctx, k, field, value).Result()
	if err != nil {
		fmt.Println("Failed to decrement age in Redis: ", err)
		return err
	}
	return nil
}

func (r *RedisClient) HGetBy(k string, field string) (interface{}, error) {
	val, err := r.client.HGet(r.ctx, k, field).Result()
	if err != nil {
		fmt.Println("Failed to get data from Redis: ", err)
		return nil, err
	}
	return val, nil
}

func (r *RedisClient) HGetALL(k string) (map[string]string, error) {
	val, err := r.client.HGetAll(r.ctx, k).Result()
	if err != nil {
		fmt.Println("Failed to get data from Redis: ", err)
		return nil, err
	}
	return val, nil
}
