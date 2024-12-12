package cache

import (
	"context"
	"errors"
	"fmt"
	"github.com/YourSuzumiya/ACMBot/app"
	"github.com/redis/go-redis/v9"
)

var (
	rdb *redis.Client
	Ctx = context.Background()
)

func init() {
	cfg := app.GetConfig().Redis
	rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := rdb.Ping(Ctx).Err(); err != nil {
		panic("redis connect failed")
	}
}

func IsNil(err error) bool {
	return errors.Is(err, redis.Nil)
}
