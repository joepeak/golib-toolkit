# Redis 缓存和分布式锁 - 简化使用示例

## 🎯 优化点 1: 无 Context 的便捷方法

### 原来的使用方式（需要 context）

```go
// ❌ 需要传入 context 的用法
ctx := context.Background()
value, err := cache.Get(ctx, "user:1")
if err != nil {
    log.Printf("Get failed: %v", err)
} else {
    if u, ok := value.(*User); ok {
        fmt.Printf("User: %+v\n", *u)
    }
}

err = cache.Set(ctx, "user:1", user, 10*time.Minute)
err = lock.Lock(ctx)
defer lock.Unlock(ctx)
```

### 优化后的使用方式（无 context）

```go
// ✅ 无 context 的用法（推荐）
client, _ := redis.NewClient(&redis.RedisConfig{
    Addr: "localhost:6379",
})

// 简单的缓存操作（自动使用 context.Background()）
value, err := client.SimpleGet("user:1")
if err != nil {
    log.Printf("Get failed: %v", err)
} else {
    // 仍需要类型断言，但代码更简洁
    if u, ok := value.(*User); ok {
        fmt.Printf("User: %+v\n", *u)
    }
}

// 设置缓存
err = client.SimpleSet("user:1", user, 10*time.Minute)

// 删除缓存
err = client.SimpleDelete("user:1")

// 检查是否存在
exists, err := client.SimpleExists("user:1")
```

## 🎯 优化点 2: 简化锁创建

### 原来的使用方式（复杂配置）

```go
// ❌ 复杂的配置方式
lock := lockMgr.NewLock("resource:123",
    redis.WithExpiration(30*time.Second),
    redis.WithRetryTimes(3),
    redis.WithAutoExtend(true),
)

err := lock.Lock(ctx)
defer lock.Unlock(ctx)
```

### 优化后的使用方式（简化）

```go
// ✅ 简化的锁创建（推荐）

// 方式1: 最简单的锁（默认配置）
lock := client.SimpleLock("resource:123")
err := lock.Lock(ctx)
defer lock.Unlock(ctx)

// 方式2: 指定超时时间
lock := client.SimpleLockWithTimeout("resource:123", 60*time.Second)
err := lock.Lock(ctx)
defer lock.Unlock(ctx)

// 方式3: 自动续期的锁
lock := client.AutoLock("resource:123")
err := lock.Lock(ctx)
defer lock.Unlock(ctx)

// 方式4: 无 context 的锁操作
err = client.SimpleLockNoCtx("resource:123")
defer client.SimpleUnlockNoCtx("resource:123")
```

## 🎯 优化点 3: 完整的最佳实践示例

### 推荐的完整使用方式

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "github.com/joepeak/golib-util/redis"
)

type User struct {
    ID    int64  `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func main() {
    // 创建客户端
    client, err := redis.NewClient(&redis.RedisConfig{
        Addr:     "localhost:6379",
        Password: "",
        DB:       0,
        PoolSize: 10,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // === 简单的缓存使用（无 context） ===
    
    // 设置用户缓存
    user := &User{ID: 1, Name: "Alice", Email: "alice@example.com"}
    err = client.SimpleSet("user:1", user, 10*time.Minute)
    if err != nil {
        log.Printf("Set failed: %v", err)
    }

    // 获取用户缓存
    value, err := client.SimpleGet("user:1")
    if err != nil {
        log.Printf("Get failed: %v", err)
    } else {
        // 仍需要类型断言，但使用更简单
        if u, ok := value.(*User); ok {
            fmt.Printf("Cached user: %+v\n", *u)
        }
    }

    // 检查是否存在
    exists, err := client.SimpleExists("user:1")
    if err != nil {
        log.Printf("Exists failed: %v", err)
    } else {
        fmt.Printf("User exists: %v\n", exists)
    }

    // === 简化的锁使用 ===
    
    // 最简单的锁
    lock := client.SimpleLock("resource:123")
    
    err = lock.Lock(context.Background())
    if err != nil {
        log.Printf("Lock failed: %v", err)
        return
    }
    defer lock.Unlock(context.Background())

    fmt.Println("Lock acquired, doing work...")
    
    // 执行业务逻辑
    time.Sleep(5 * time.Second)
    
    fmt.Println("Work completed, releasing lock")

    // === 无 context 的锁操作 ===
    
    err = client.SimpleLockNoCtx("resource:456")
    if err != nil {
        log.Printf("SimpleLockNoCtx failed: %v", err)
    } else {
        fmt.Println("Simple lock acquired")
        defer client.SimpleUnlockNoCtx("resource:456")
    }

    // === 高级功能（仍需 context） ===
    
    // 如果需要高级功能，仍可使用原始接口
    cache := client.Cache()
    
    // 带自动加载的获取
    userLoader := func(ctx context.Context, key string) (*any, error) {
        log.Printf("Loading user from database: %s", key)
        return &User{ID: 1, Name: "Alice", Email: "alice@example.com"}, nil
    }

    loadedValue, err := cache.GetOrLoad(context.Background(), "user:2", userLoader)
    if err != nil {
        log.Printf("GetOrLoad failed: %v", err)
    } else {
        if u, ok := loadedValue.(*User); ok {
            fmt.Printf("Loaded user: %+v\n", *u)
        }
    }

    // === 查看指标 ===
    metrics := cache.Metrics()
    stats := metrics.GetStats()
    
    fmt.Printf("Cache Stats:\n")
    fmt.Printf("  Hits: %d\n", stats.Hits)
    fmt.Printf("  Misses: %d\n", stats.Misses)
    fmt.Printf("  Hit Rate: %.2f%%\n", stats.HitRate)
}
```

## 🎯 总结优化效果

### 优化前的问题
1. ❌ **强制使用 context**: 简单操作也需要传入 context
2. ❌ **锁配置复杂**: 需要了解很多配置选项
3. ❌ **代码冗长**: 每次都要写配置选项

### 优化后的优势
1. ✅ **简化使用**: `SimpleGet()` 自动处理 context
2. ✅ **简化锁**: `SimpleLock()` 一行搞定
3. ✅ **保持灵活**: 高级功能仍可使用原始接口
4. ✅ **渐进式**: 可以从简单开始，逐步使用高级功能

### 推荐的使用模式

```go
// 1. 创建客户端
client, _ := redis.NewClient(config)

// 2. 简单操作（无 context）
value, err := client.SimpleGet("key")
err = client.SimpleSet("key", value, ttl)
err = client.SimpleDelete("key")

// 3. 简单锁
lock := client.SimpleLock("resource")
err = lock.Lock(ctx)
defer lock.Unlock(ctx)

// 4. 高级功能（需要 context）
cache := client.Cache()
value, err := cache.GetOrLoad(ctx, key, loader)
```

## 🎯 API 对比

| 功能 | 原来方式 | 优化后方式 |
|------|----------|------------|
| 获取缓存 | `cache.Get(ctx, key)` | `client.SimpleGet(key)` |
| 设置缓存 | `cache.Set(ctx, key, value, ttl)` | `client.SimpleSet(key, value, ttl)` |
| 删除缓存 | `cache.Delete(ctx, key)` | `client.SimpleDelete(key)` |
| 检查存在 | `cache.Exists(ctx, key)` | `client.SimpleExists(key)` |
| 创建锁 | `lockMgr.NewLock(key, opts...)` | `client.SimpleLock(key)` |
| 锁操作 | `lock.Lock(ctx)` | `client.SimpleLockNoCtx(key)` |

这样的设计既保持了灵活性，又大大简化了常见操作的使用！🎉
