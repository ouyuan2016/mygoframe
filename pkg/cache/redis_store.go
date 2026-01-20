package cache

import (
	"context"
	"errors"
	"fmt"
	"mygoframe/pkg/config"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	client *redis.Client
	prefix string
}

func NewRedisRepository(cfg *config.Config) (*RedisRepository, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		MaxRetries:   cfg.Redis.MaxRetries,
		DialTimeout:  time.Duration(cfg.Redis.DialTimeout) * time.Second,
		ReadTimeout:  time.Duration(cfg.Redis.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Redis.WriteTimeout) * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("连接Redis失败: %w", err)
	}

	return &RedisRepository{
		client: client,
		prefix: cfg.Redis.Prefix,
	}, nil
}

func (r *RedisRepository) Get(ctx context.Context, key string) (string, error) {
	fullKey := r.prefix + key
	val, err := r.client.Get(ctx, fullKey).Result()
	if err == redis.Nil {
		return "", errors.New("缓存不存在")
	}
	if err != nil {
		return "", err
	}
	return val, nil
}

func (r *RedisRepository) Put(ctx context.Context, key string, value string, ttl time.Duration) error {
	fullKey := r.prefix + key
	return r.client.Set(ctx, fullKey, value, ttl).Err()
}

func (r *RedisRepository) Forget(ctx context.Context, key string) error {
	fullKey := r.prefix + key
	return r.client.Del(ctx, fullKey).Err()
}

func (r *RedisRepository) Has(ctx context.Context, key string) (bool, error) {
	fullKey := r.prefix + key
	exists, err := r.client.Exists(ctx, fullKey).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func (r *RedisRepository) Flush(ctx context.Context) error {
	var cursor uint64
	var keys []string

	for {
		var scanKeys []string
		var err error
		scanKeys, cursor, err = r.client.Scan(ctx, cursor, r.prefix+"*", 100).Result()
		if err != nil {
			return err
		}

		keys = append(keys, scanKeys...)

		if cursor == 0 {
			break
		}
	}

	if len(keys) > 0 {
		return r.client.Del(ctx, keys...).Err()
	}

	return nil
}

func (r *RedisRepository) Close() error {
	return r.client.Close()
}

func (r *RedisRepository) Client() interface{} {
	return r.client
}
