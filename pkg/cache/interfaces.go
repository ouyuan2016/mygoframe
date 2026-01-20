package cache

import (
	"context"
	"time"
)

type Repository interface {
	Get(ctx context.Context, key string) (string, error)
	Put(ctx context.Context, key string, value string, ttl time.Duration) error
	Forget(ctx context.Context, key string) error
	Has(ctx context.Context, key string) (bool, error)
	Flush(ctx context.Context) error
	Client() interface{}
}

type Cache interface {
	Repository
	PutObject(ctx context.Context, key string, obj interface{}, ttl time.Duration) error
	GetObject(ctx context.Context, key string, obj interface{}) error
	Close() error
}
