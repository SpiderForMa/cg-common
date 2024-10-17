package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"log"
)

var RedisClient *redis.Client
var ctx = context.Background()

// RedisConfig 用于存储 Redis 配置
type RedisConfig struct {
	Addr     string `json:"addr"`     // Redis 地址
	Password string `json:"password"` // Redis 密码
	DB       int    `json:"db"`       // 使用的数据库编号
}

// InitRedis 初始化 Redis 连接
func InitRedis(config RedisConfig) {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	// 测试连接
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
	}

	log.Println("Connected to Redis successfully!")
}

// GetRedisClient 返回 Redis 客户端
func GetRedisClient() *redis.Client {
	return RedisClient
}
