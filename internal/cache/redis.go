package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gotracker/internal/order"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisCache(addr string, ttlSeconds int) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	return &RedisCache{
		client: client,
		ttl:    time.Duration(ttlSeconds) * time.Second,
	}
}

func (c *RedisCache) key(id int) string {
	return fmt.Sprintf("order:%d", id)
}

func (c *RedisCache) Get(id int) (order.Order, error) {
	ctx := context.Background()

	data, err := c.client.Get(
		ctx,
		c.key(id),
	).Result()

	if err == redis.Nil {
		return order.Order{}, fmt.Errorf("not found in cache")
	}

	if err != nil {
		return order.Order{}, err
	}

	var o order.Order

	if err := json.Unmarshal([]byte(data), &o); err != nil {
		return order.Order{}, err
	}

	return o, nil
}

func (c *RedisCache) Set(o order.Order) error {
	ctx := context.Background()

	data, err := json.Marshal(o)
	if err != nil {
		return err
	}

	return c.client.Set(
		ctx,
		c.key(o.ID),
		data,
		c.ttl,
	).Err()
}

func (c *RedisCache) Delete(id int) error {
	ctx := context.Background()

	return c.client.Del(
		ctx,
		c.key(id),
	).Err()
}
