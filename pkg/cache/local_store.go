package cache

import (
	"context"
	"errors"
	"fmt"
	"mygoframe/pkg/config"
	"time"

	"github.com/dgraph-io/ristretto"
)

type LocalRepository struct {
	cache  *ristretto.Cache
	prefix string
}

func NewLocalRepository(cfg *config.Config) (*LocalRepository, error) {
	maxCost := cfg.LocalCache.MaxCost
	if maxCost == 0 {
		maxCost = 1 << 30
	}

	maxKeys := cfg.LocalCache.MaxKeys
	if maxKeys == 0 {
		maxKeys = 1e6
	}

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: maxKeys,
		MaxCost:     maxCost,
		BufferItems: 64,
	})
	if err != nil {
		return nil, fmt.Errorf("创建本地缓存失败: %w", err)
	}

	return &LocalRepository{
		cache:  cache,
		prefix: "",
	}, nil
}

func (l *LocalRepository) Get(ctx context.Context, key string) (string, error) {
	fullKey := l.prefix + key
	val, found := l.cache.Get(fullKey)
	if !found {
		return "", errors.New("缓存不存在")
	}

	strVal, ok := val.(string)
	if !ok {
		return "", errors.New("缓存值类型错误")
	}

	return strVal, nil
}

func (l *LocalRepository) Put(ctx context.Context, key string, value string, ttl time.Duration) error {
	fullKey := l.prefix + key
	cost := int64(len(value))

	if ttl > 0 {
		l.cache.SetWithTTL(fullKey, value, cost, ttl)
	} else {
		l.cache.Set(fullKey, value, cost)
	}

	l.cache.Wait()
	return nil
}

func (l *LocalRepository) Forget(ctx context.Context, key string) error {
	fullKey := l.prefix + key
	l.cache.Del(fullKey)
	return nil
}

func (l *LocalRepository) Has(ctx context.Context, key string) (bool, error) {
	fullKey := l.prefix + key
	_, found := l.cache.Get(fullKey)
	return found, nil
}

func (l *LocalRepository) Flush(ctx context.Context) error {
	maxCost := int64(1 << 30)
	maxCounters := int64(1e6)

	newCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: maxCounters,
		MaxCost:     maxCost,
		BufferItems: 64,
	})
	if err != nil {
		return fmt.Errorf("重新创建本地缓存失败: %w", err)
	}

	l.cache = newCache
	return nil
}

func (l *LocalRepository) Close() error {
	if l.cache != nil {
		l.cache.Close()
	}
	return nil
}

func (l *LocalRepository) Client() interface{} {
	return l.cache
}
