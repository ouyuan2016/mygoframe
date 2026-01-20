package cache

import (
	"context"
	"errors"
	"mygoframe/pkg/config"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/redis/go-redis/v9"
)

var globalStore *Manager

func Init(cfg *config.Config) error {
	manager := NewManager(cfg)
	if err := manager.Init(); err != nil {
		return err
	}

	globalStore = manager
	return nil
}

func Close() error {
	if globalStore != nil {
		return globalStore.Close()
	}
	return nil
}

func ensureInitialized() error {
	if globalStore == nil {
		return errors.New("缓存系统未初始化")
	}
	return nil
}

func Get(ctx context.Context, key string) (string, error) {
	if err := ensureInitialized(); err != nil {
		return "", err
	}
	return globalStore.Get(ctx, key)
}

func Put(ctx context.Context, key string, value string, ttl time.Duration) error {
	if err := ensureInitialized(); err != nil {
		return err
	}
	return globalStore.Put(ctx, key, value, ttl)
}

func Forget(ctx context.Context, key string) error {
	if err := ensureInitialized(); err != nil {
		return err
	}
	return globalStore.Forget(ctx, key)
}

func Has(ctx context.Context, key string) (bool, error) {
	if err := ensureInitialized(); err != nil {
		return false, err
	}
	return globalStore.Has(ctx, key)
}

func Flush(ctx context.Context) error {
	if err := ensureInitialized(); err != nil {
		return err
	}
	return globalStore.Flush(ctx)
}

func PutObject(ctx context.Context, key string, obj interface{}, ttl time.Duration) error {
	if err := ensureInitialized(); err != nil {
		return err
	}
	return globalStore.PutObject(ctx, key, obj, ttl)
}

func GetObject(ctx context.Context, key string, obj interface{}) error {
	if err := ensureInitialized(); err != nil {
		return err
	}
	return globalStore.GetObject(ctx, key, obj)
}

// Store 获取指定名称的缓存存储
func Store(storeName string) Repository {
	if globalStore == nil {
		return nil
	}
	return globalStore.Store(storeName)
}

// Local 获取本地缓存存储
func Local() Repository {
	return Store("local")
}

// Redis 获取Redis缓存存储
func Redis() Repository {
	return Store("redis")
}

// GetClient 获取指定缓存存储的底层客户端
func GetClient(storeName string) interface{} {
	store := Store(storeName)
	if store == nil {
		return nil
	}
	return store.Client()
}

// GetLocalClient 获取本地缓存的底层客户端 (Ristretto)
func GetLocalClient() *ristretto.Cache {
	client := GetClient("local")
	if client == nil {
		return nil
	}
	if cache, ok := client.(*ristretto.Cache); ok {
		return cache
	}
	return nil
}

// GetRedisClient 获取Redis缓存的底层客户端 (redis.Client)
func GetRedisClient() *redis.Client {
	client := GetClient("redis")
	if client == nil {
		return nil
	}
	if client, ok := client.(*redis.Client); ok {
		return client
	}
	return nil
}
