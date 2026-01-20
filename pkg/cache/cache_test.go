package cache

import (
	"context"
	"testing"
	"time"

	"mygoframe/pkg/config"
)

func TestCacheBasicOperations(t *testing.T) {
	cfg := &config.Config{
		Redis: config.Redis{
			Enabled: false, // 使用本地缓存测试
		},
		LocalCache: config.LocalCache{
			MaxCost: 1 << 30,
			MaxKeys: 1e6,
		},
	}

	// 初始化缓存
	err := Init(cfg)
	if err != nil {
		t.Fatalf("初始化缓存失败: %v", err)
	}
	defer Close()

	ctx := context.Background()
	key := "test_key"
	value := "test_value"

	// 测试Put
	err = Put(ctx, key, value, 5*time.Minute)
	if err != nil {
		t.Errorf("Put失败: %v", err)
	}

	// 测试Has
	exists, err := Has(ctx, key)
	if err != nil {
		t.Errorf("Has失败: %v", err)
	}
	if !exists {
		t.Error("Key应该存在")
	}

	// 测试Get
	retrievedValue, err := Get(ctx, key)
	if err != nil {
		t.Errorf("Get失败: %v", err)
	}
	if retrievedValue != value {
		t.Errorf("获取的值不匹配: 期望 %s, 实际 %s", value, retrievedValue)
	}

	// 测试Forget
	err = Forget(ctx, key)
	if err != nil {
		t.Errorf("Forget失败: %v", err)
	}

	// 验证删除
	exists, err = Has(ctx, key)
	if err != nil {
		t.Errorf("Has失败: %v", err)
	}
	if exists {
		t.Error("Key不应该存在")
	}
}

func TestCacheObjectOperations(t *testing.T) {
	cfg := &config.Config{
		Redis: config.Redis{
			Enabled: false,
		},
		LocalCache: config.LocalCache{
			MaxCost: 1 << 30,
			MaxKeys: 1e6,
		},
	}

	err := Init(cfg)
	if err != nil {
		t.Fatalf("初始化缓存失败: %v", err)
	}
	defer Close()

	ctx := context.Background()

	type TestUser struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	user := TestUser{
		ID:   1,
		Name: "张三",
	}

	// 测试对象存储
	err = PutObject(ctx, "user:1", user, 5*time.Minute)
	if err != nil {
		t.Errorf("PutObject失败: %v", err)
	}

	// 测试对象获取
	var retrievedUser TestUser
	err = GetObject(ctx, "user:1", &retrievedUser)
	if err != nil {
		t.Errorf("GetObject失败: %v", err)
	}

	if retrievedUser.ID != user.ID || retrievedUser.Name != user.Name {
		t.Errorf("获取的对象不匹配: 期望 %+v, 实际 %+v", user, retrievedUser)
	}
}
