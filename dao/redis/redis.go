package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"web_app/settings"
)

// 声明一个全局的rdb变量
var rdb *redis.Client

// Init 初始化Redis连接
func Init(cfg *settings.RedisConfig) (err error) {
	rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx := context.Background()
	_, err = rdb.Ping(ctx).Result()
	if err != nil {
		return
	}
	return
}

func Close() {
	_ = rdb.Close()
}
