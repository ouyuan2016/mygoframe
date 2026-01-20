package cache

import (
	"context"
	"encoding/json"
	"fmt"
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

func SerializeObject(obj interface{}) (string, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("序列化对象失败: %w", err)
	}
	return string(data), nil
}

func DeserializeObject(data string, obj interface{}) error {
	return json.Unmarshal([]byte(data), obj)
}
