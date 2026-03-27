# Redis 缓存和分布式锁使用示例

## 基础缓存使用

```go
package main

import (
    "context"
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
    // 创建 Redis 客户端
    config := &redis.RedisConfig{
        Addr:     "localhost:6379",
        Password: "",
        DB:       0,
        PoolSize: 10,
    }

    client, err := redis.NewClient(config)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    ctx := context.Background()
    cache := client.Cache()

    // 基础缓存操作
    user := &User{ID: 1, Name: "Alice", Email: "alice@example.com"}
    
    // 设置缓存
    err = cache.Set(ctx, "user:1", user, 10*time.Minute)
    if err != nil {
        log.Printf("Set failed: %v", err)
    }

    // 获取缓存
    var cachedUser User
    value, err := cache.Get(ctx, "user:1")
    if err != nil {
        log.Printf("Get failed: %v", err)
    } else {
        // 类型断言
        if u, ok := value.(*User); ok {
            cachedUser = *u
            fmt.Printf("Cached user: %+v\n", cachedUser)
        }
    }

    // 批量操作
    users := map[string]*User{
        "user:2": {ID: 2, Name: "Bob", Email: "bob@example.com"},
        "user:3": {ID: 3, Name: "Charlie", Email: "charlie@example.com"},
    }
    err = cache.SetMany(ctx, users, 5*time.Minute)
    if err != nil {
        log.Printf("SetMany failed: %v", err)
    }

    keys := []string{"user:1", "user:2", "user:3"}
    cachedUsers, err := cache.GetMany(ctx, keys)
    if err != nil {
        log.Printf("GetMany failed: %v", err)
    } else {
        for key, user := range cachedUsers {
            fmt.Printf("%s: %+v\n", key, user)
        }
    }
}
```

## 高级缓存使用（自动加载）

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/joepeak/golib-util/redis"
)

func main() {
    config := &redis.RedisConfig{
        Addr: "localhost:6379",
    }

    client, err := redis.NewClient(config)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    ctx := context.Background()
    cache := client.Cache()

    // 带自动加载的获取
    userLoader := func(ctx context.Context, key string) (*User, error) {
        log.Printf("Loading user from database: %s", key)
        // 模拟数据库查询
        return &User{ID: 1, Name: "Alice", Email: "alice@example.com"}, nil
    }

    user, err := cache.GetOrLoad(ctx, "user:1", userLoader)
    if err != nil {
        log.Printf("GetOrLoad failed: %v", err)
    } else {
        fmt.Printf("Loaded user: %+v\n", user)
    }

    // 批量加载
    batchLoader := func(ctx context.Context, keys []string) (map[string]*User, error) {
        log.Printf("Batch loading users: %v", keys)
        result := make(map[string]*User)
        for _, key := range keys {
            result[key] = &User{ID: 2, Name: "Bob", Email: "bob@example.com"}
        }
        return result, nil
    }

    keys := []string{"user:2", "user:3"}
    users, err := cache.GetManyOrLoad(ctx, keys, batchLoader)
    if err != nil {
        log.Printf("GetManyOrLoad failed: %v", err)
    } else {
        for key, user := range users {
            fmt.Printf("%s: %+v\n", key, user)
        }
    }
}
```

## 分布式锁使用

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/joepeak/golib-util/redis"
)

func main() {
    config := &redis.RedisConfig{
        Addr: "localhost:6379",
    }

    client, err := redis.NewClient(config)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    ctx := context.Background()
    lockMgr := client.LockManager()

    // 创建锁
    lock := lockMgr.NewLock("resource:123",
        redis.WithExpiration(30*time.Second),
        redis.WithRetryTimes(3),
        redis.WithAutoExtend(true),
    )

    // 获取锁
    err = lock.Lock(ctx)
    if err != nil {
        log.Printf("Lock failed: %v", err)
        return
    }
    defer lock.Unlock(ctx)

    fmt.Println("Lock acquired, doing work...")

    // 执行业务逻辑
    time.Sleep(10 * time.Second)

    fmt.Println("Work completed, releasing lock")
}
```

## 批量锁操作

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/joepeak/golib-util/redis"
)

func main() {
    config := &redis.RedisConfig{
        Addr: "localhost:6379",
    }

    client, err := redis.NewClient(config)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    ctx := context.Background()
    lockMgr := client.LockManager()

    // 批量获取锁
    keys := []string{"resource:1", "resource:2", "resource:3"}
    locks, err := lockMgr.LockMany(ctx, keys,
        redis.WithExpiration(30*time.Second),
        redis.WithRetryTimes(3),
    )
    if err != nil {
        log.Printf("LockMany failed: %v", err)
        return
    }

    fmt.Printf("Acquired %d locks\n", len(locks))

    // 执行业务逻辑
    time.Sleep(5 * time.Second)

    // 批量释放锁
    err = lockMgr.UnlockMany(ctx, locks)
    if err != nil {
        log.Printf("UnlockMany failed: %v", err)
    } else {
        fmt.Println("All locks released")
    }
}
```

## 配置选项

### 缓存配置

```go
// 创建带配置的缓存
cache := redis.NewRedisCache[any](client,
    redis.WithDefaultTTL(10*time.Minute),
    redis.WithNullValueTTL(1*time.Minute),
    redis.WithEnableSingleFlight(true),
    redis.WithEnableMetrics(true),
    redis.WithEnableHotKeyDetect(true),
    redis.WithHotKeyThreshold(100),
)
```

### 锁配置

```go
// 创建带配置的锁
lock := redis.NewRedisDistributedLock(client, "my-lock",
    redis.WithExpiration(30*time.Second),
    redis.WithRetryTimes(5),
    redis.WithRetryDelay(200*time.Millisecond),
    redis.WithAutoExtend(true),
    redis.WithExtendBefore(5*time.Second),
)
```

## 从 Viper 配置创建客户端

```go
package main

import (
    "log"
    
    "github.com/joepeak/golib-util/redis"
    _ "github.com/joepeak/golib-conf" // 导入配置
)

func main() {
    // 从配置文件创建客户端
    // 配置文件示例：
    // redis:
    //   addr: "localhost:6379"
    //   password: ""
    //   db: 0
    //   enabledCluster: false
    //   poolSize: 10
    
    client, err := redis.NewClientFromViper("redis")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // 使用客户端...
    cache := client.Cache()
    lockMgr := client.LockManager()
    
    // ...
}
```

## 指标统计

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/joepeak/golib-util/redis"
)

func main() {
    config := &redis.RedisConfig{
        Addr: "localhost:6379",
    }

    client, err := redis.NewClient(config)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    cache := client.Cache()

    // 执行一些操作...
    cache.Set(context.Background(), "key1", "value1", time.Minute)
    cache.Get(context.Background(), "key1")
    cache.Get(context.Background(), "key2") // miss

    // 获取指标
    metrics := cache.Metrics()
    stats := metrics.GetStats()

    fmt.Printf("Cache Stats:\n")
    fmt.Printf("  Hits: %d\n", stats.Hits)
    fmt.Printf("  Misses: %d\n", stats.Misses)
    fmt.Printf("  Hit Rate: %.2f%%\n", stats.HitRate)
    fmt.Printf("  Total Operations: %d\n", stats.TotalOperations)

    if stats.GetLatency != nil {
        fmt.Printf("  Get Latency - Avg: %v, P95: %v, P99: %v\n",
            stats.GetLatency.Avg,
            stats.GetLatency.P95,
            stats.GetLatency.P99)
    }

    if len(stats.HotKeys) > 0 {
        fmt.Printf("  Hot Keys: %v\n", stats.HotKeys)
    }
}
```

## 集群配置

```go
package main

import (
    "log"
    
    "github.com/joepeak/golib-util/redis"
)

func main() {
    // 集群配置
    config := &redis.RedisConfig{
        EnabledCluster: true,
        ClusterAddrs: []string{
            "redis-node1:6379",
            "redis-node2:6379",
            "redis-node3:6379",
        },
        Password: "your-password",
        PoolSize: 20,
    }

    client, err := redis.NewClient(config)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // 使用集群客户端...
    cache := client.Cache()
    lockMgr := client.LockManager()
    
    // ...
}
```
