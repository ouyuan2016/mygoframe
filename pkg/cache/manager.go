package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"mygoframe/pkg/config"
	"mygoframe/pkg/logger"
	"time"

	"go.uber.org/zap"
)

type Manager struct {
	config       *config.Config
	stores       map[string]Repository
	defaultStore string
}

func NewManager(cfg *config.Config) *Manager {
	return &Manager{
		config:       cfg,
		stores:       make(map[string]Repository),
		defaultStore: "",
	}
}

// GetStore 获取指定名称的缓存存储
func (m *Manager) GetStore(name string) Repository {
	return m.Store(name)
}

// Store 获取指定名称的缓存存储（合并实现）
func (m *Manager) Store(name string) Repository {
	if store, exists := m.stores[name]; exists {
		return store
	}
	// 回退到本地缓存
	return m.stores["local"]
}

func (m *Manager) Init() error {
	localRepo, err := NewLocalRepository(m.config)
	if err != nil {
		return fmt.Errorf("初始化本地缓存失败: %w", err)
	}
	m.stores["local"] = localRepo

	if m.config.Redis.Enabled {
		m.defaultStore = "redis"

		redisRepo, err := NewRedisRepository(m.config)
		if err != nil {
			logger.Warn("Redis缓存初始化失败，降级到本地缓存", zap.Error(err))
			m.defaultStore = "local"
			return nil
		}
		m.stores["redis"] = redisRepo
	} else {
		m.defaultStore = "local"
	}

	return nil
}

func (m *Manager) Get(ctx context.Context, key string) (string, error) {
	return m.Store(m.defaultStore).Get(ctx, key)
}

func (m *Manager) Put(ctx context.Context, key string, value string, ttl time.Duration) error {
	return m.Store(m.defaultStore).Put(ctx, key, value, ttl)
}

func (m *Manager) Forget(ctx context.Context, key string) error {
	return m.Store(m.defaultStore).Forget(ctx, key)
}

func (m *Manager) Has(ctx context.Context, key string) (bool, error) {
	return m.Store(m.defaultStore).Has(ctx, key)
}

func (m *Manager) Flush(ctx context.Context) error {
	return m.Store(m.defaultStore).Flush(ctx)
}

func (m *Manager) PutObject(ctx context.Context, key string, obj interface{}, ttl time.Duration) error {
	data, err := SerializeObject(obj)
	if err != nil {
		return err
	}
	return m.Put(ctx, key, data, ttl)
}

func (m *Manager) GetObject(ctx context.Context, key string, obj interface{}) error {
	data, err := m.Get(ctx, key)
	if err != nil {
		return err
	}
	return DeserializeObject(data, obj)
}

func (m *Manager) Close() error {
	var errs []error

	if redisStore, exists := m.stores["redis"]; exists {
		if closer, ok := redisStore.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}

	if localStore, exists := m.stores["local"]; exists {
		if closer, ok := localStore.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("关闭缓存存储失败: %v", errs)
	}
	return nil
}

// Client 返回默认缓存存储的底层客户端
func (m *Manager) Client() interface{} {
	return m.Store(m.defaultStore).Client()
}

func SerializeObject(obj interface{}) (string, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("序列化对象失败: %w", err)
	}
	return string(data), nil
}

func DeserializeObject(data string, obj interface{}) error {
	// 检查数据是否为空
	if data == "" {
		return nil
	}
	return json.Unmarshal([]byte(data), obj)
}
