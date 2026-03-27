# Golib Util

Go 语言工具集库，提供数据库、对象存储、加密、验证等常用工具函数和组件。

## 功能特性

- 🗄️ **数据库支持**: MySQL、PostgreSQL、MongoDB 等主流数据库
- ☁️ **对象存储**: 阿里云 OSS、腾讯云 COS、AWS S3 等云存储
-  **加密工具**: 哈希、加密、编码解码等安全工具
- 🎲 **随机生成**: 雪花算法、UUID、NanoID 等唯一 ID 生成
- 🌍 **国际化**: i18n 多语言支持
- 📧 **工具函数**: 字符串、数字、时间、类型转换等常用函数
- 📬 **邮件服务**: SMTP 邮件发送
- 🌐 **RPC 支持**: gRPC 客户端封装
- 🔄 **单例模式**: 线程安全的单例工具

## 快速开始

### 安装

```bash
# 安装完整工具集
go get github.com/joepeak/golib-util

# 或通过元包安装
go get github.com/joepeak/golib-toolbox
```

### 数据库使用

```go
package main

import (
    "context"
    "log"
    
    "github.com/joepeak/golib-util/database"
    _ "github.com/joepeak/golib-conf"  // 配置初始化
)

func main() {
    // MySQL 连接
    mysqlDB, err := database.NewMySQL(&database.MySQLConfig{
        Host:     "localhost",
        Port:     3306,
        Database:  "testdb",
        Username:  "root",
        Password:  "password",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // 执行查询
    var users []User
    err = mysqlDB.Where("age > ?", 18).Find(&users).Error
    if err != nil {
        log.Printf("查询失败: %v", err)
    }
    
    // PostgreSQL 连接
    pgDB, err := database.NewPostgreSQL(&database.PgSQLConfig{
        Host:     "localhost",
        Port:     5432,
        Database:  "testdb",
        Username:  "postgres",
        Password:  "password",
    })
    
    // MongoDB 连接
    mongoDB, err := database.NewMongoDB(&database.MongoConfig{
        URI:      "mongodb://localhost:27017",
        Database:  "testdb",
    })
}
```

### 对象存储使用

```go
package main

import (
    "context"
    "log"
    "os"
    
    "github.com/joepeak/golib-util/oss"
    _ "github.com/joepeak/golib-conf"
)

func main() {
    // 阿里云 OSS
    aliyunOSS, err := oss.NewAliyunOSS(&oss.AliyunConfig{
        AccessKeyID:     "your-access-key-id",
        AccessKeySecret:  "your-access-key-secret",
        Bucket:          "your-bucket",
        Endpoint:        "oss-cn-hangzhou.aliyuncs.com",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // 上传文件
    file, _ := os.Open("test.txt")
    err = aliyunOSS.Upload("test.txt", file)
    if err != nil {
        log.Printf("上传失败: %v", err)
    }
    
    // AWS S3
    s3, err := oss.NewAWSS3(&oss.AWSConfig{
        AccessKeyID:     "your-access-key-id",
        AccessKeySecret:  "your-access-key-secret",
        Bucket:          "your-bucket",
        Region:          "us-west-2",
    })
    
    // 腾讯云 COS
    cos, err := oss.NewTxyunCOS(&oss.TxyunConfig{
        SecretID:  "your-secret-id",
        SecretKey: "your-secret-key",
        Bucket:    "your-bucket",
        Region:    "ap-guangzhou",
    })
}
```

### Redis 缓存使用

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/joepeak/golib-util/redisclient"
    _ "github.com/joepeak/golib-conf"
)

func main() {
    // 创建 Redis 客户端
    client := redisclient.NewClient(&redisclient.RedisConfig{
        Addr:     "localhost:6379",
        Password: "",
        DB:       0,
        PoolSize:  10,
    })
    
    // 设置缓存
    err := client.Set("user:123", "张三", time.Hour)
    if err != nil {
        log.Printf("设置缓存失败: %v", err)
    }
    
    // 获取缓存
    var name string
    err = client.Get("user:123", &name)
    if err != nil {
        log.Printf("获取缓存失败: %v", err)
    } else {
        log.Printf("用户名: %s", name)
    }
    
    // 分布式锁
    lock := redisclient.NewLock("user:123:lock", time.Minute*5)
    err = lock.Lock()
    if err != nil {
        log.Printf("获取锁失败: %v", err)
        return
    }
    
    defer lock.Unlock()
    
    // 执行需要加锁的操作
    log.Println("执行业务逻辑...")
}
```

### 工具函数使用

```go
package main

import (
    "fmt"
    
    "github.com/joepeak/golib-util/convert"
    "github.com/joepeak/golib-util/rand"
    "github.com/joepeak/golib-util/snowflake"
    "github.com/joepeak/golib-util/nanoid"
    "github.com/joepeak/golib-util/verify"
)

func main() {
    // 字符串转换
    str := "Hello, World!"
    reversed := convert.Reverse(str)
    upper := convert.ToUpper(str)
    
    // 随机数生成
    randomInt := rand.Int(1000)
    randomString := rand.String(10)
    
    // 唯一 ID 生成
    snowflakeID := snowflake.NextID()
    nanoid := nanoid.Generate()
    
    // 数据验证
    email := "user@example.com"
    isValid := verify.IsEmail(email)
    phone := "13812345678"
    maskedPhone := verify.MaskPhone(phone)
    
    fmt.Printf("原字符串: %s\n", str)
    fmt.Printf("反转后: %s\n", reversed)
    fmt.Printf("大写: %s\n", upper)
    fmt.Printf("随机数: %d\n", randomInt)
    fmt.Printf("随机字符串: %s\n", randomString)
    fmt.Printf("Snowflake ID: %d\n", snowflakeID)
    fmt.Printf("NanoID: %s\n", nanoid)
    fmt.Printf("邮箱有效: %v\n", isValid)
    fmt.Printf("手机号脱敏: %s\n", maskedPhone)
}
```

## 配置

### 数据库配置

```yaml
database:
  mysql:
    host: "localhost"
    port: 3306
    database: "myapp"
    username: "root"
    password: "password"
    charset: "utf8mb4"
    
  postgresql:
    host: "localhost"
    port: 5432
    database: "myapp"
    username: "postgres"
    password: "password"
    sslmode: "disable"
    
  mongodb:
    uri: "mongodb://localhost:27017"
    database: "myapp"
    timeout: 30s
```

### 对象存储配置

```yaml
oss:
  aliyun:
    access_key_id: "your-access-key-id"
    access_key_secret: "your-access-key-secret"
    bucket: "your-bucket"
    endpoint: "oss-cn-hangzhou.aliyuncs.com"
    
  aws:
    access_key_id: "your-access-key-id"
    access_key_secret: "your-access-key-secret"
    bucket: "your-bucket"
    region: "us-west-2"
    
  txyun:
    secret_id: "your-secret-id"
    secret_key: "your-secret-key"
    bucket: "your-bucket"
    region: "ap-guangzhou"
```

## 项目结构

```
golib-util/
├── convert/           # 数据转换工具
│   ├── convert.go
│   ├── encode.go
│   ├── math.go
│   ├── string.go
│   └── time.go
├── database/          # 数据库客户端
│   ├── mysql.go
│   ├── pqsql.go
│   └── mongo.go
├── oss/              # 对象存储
│   ├── aliyun.go
│   ├── aws.go
│   ├── local.go
│   ├── oss.go
│   └── txyun.go
├── redisclient/       # Redis 客户端
│   ├── locker.go
│   ├── pubsub.go
│   └── redisclient.go
├── rand/             # 随机数生成
│   └── rand.go
├── snowflake/         # 雪花算法
│   └── snowflake.go
├── nanoid/           # NanoID 生成
│   └── nanoid.go
├── verify/           # 数据验证
│   └── verify.go
├── i18n/            # 国际化
│   └── i18n.go
├── smtp/             # 邮件服务
│   └── smtp.go
├── rpc/              # RPC 客户端
│   └── rpc.go
├── merkle/           # 默克尔树
│   └── merkle.go
├── structure/        # 结构体工具
│   ├── enter.go
│   └── type.go
├── main.go
├── go.mod
├── go.sum
└── README.md
```

## 依赖

### 数据库
- `gorm.io/gorm` - ORM 框架
- `gorm.io/driver/mysql` - MySQL 驱动
- `gorm.io/driver/postgres` - PostgreSQL 驱动
- `go.mongodb.org/mongo-driver/v2` - MongoDB 驱动

### 对象存储
- `github.com/aliyun/aliyun-oss-go-sdk` - 阿里云 OSS
- `github.com/aws/aws-sdk-go` - AWS S3
- `github.com/tencentyun/cos-go-sdk-v5` - 腾讯云 COS

### 缓存
- `github.com/redis/go-redis/v9` - Redis 客户端
- `github.com/go-redsync/redsync/v4` - 分布式锁

### 其他工具
- `github.com/matoous/go-nanoid/v2` - NanoID 生成
- `github.com/sony/sonyflake` - 雪花算法
- `github.com/nicksnyder/go-i18n/v2` - 国际化
- `gopkg.in/gomail.v2` - 邮件发送

## 性能特点

- **数据库**: 连接池管理，支持读写分离
- **对象存储**: 断点续传，并发上传
- **缓存**: 集群支持，自动重连
- **工具函数**: 高性能算法，内存优化

## 最佳实践

1. **连接管理**: 合理设置连接池大小
2. **错误处理**: 实现重试和降级机制
3. **资源释放**: 及时释放数据库连接和文件句柄
4. **配置管理**: 使用环境变量管理敏感信息

## 示例项目

查看 `examples/` 目录：

- [数据库 CRUD 操作](examples/database-crud/)
- [文件上传下载](examples/oss-upload/)
- [缓存使用模式](examples/redis-patterns/)
- [工具函数集合](examples/utils-collection/)

## 贡献

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建特性分支: `git checkout -b feature/amazing-feature`
3. 提交更改: `git commit -m 'Add amazing feature'`
4. 推送分支: `git push origin feature/amazing-feature`
5. 提交 Pull Request

## 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

## 作者

[@joepeak](https://github.com/joepeak)

## 更新日志

### v0.3.0
- ✨ 新增 MongoDB 支持
- 🔧 优化 Redis 连接池
- 📝 完善工具函数文档

### v0.2.0
- 🎉 初始版本发布
- 📦 数据库和 OSS 基础功能
