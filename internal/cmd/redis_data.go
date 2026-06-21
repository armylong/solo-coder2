package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/armylong/armylong-go/internal/business/demo"
	"github.com/armylong/armylong-go/internal/common/webcache"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cast"
	"github.com/urfave/cli/v2"
)

// 查询Redis缓存数据
func GetRedisData(c *cli.Context) error {
	ctx := c.Context

	cacheKey := c.Args().Get(0)

	cacheValue, _ := webcache.RedisClient.Get(ctx, cacheKey).Result()

	fmt.Printf("cache_key: %s, cache_value: %s\n", cacheKey, cacheValue)
	return nil
}

func RedisHandler(c *cli.Context) error {
	res, err := demo.DemoBusiness.SetMessage(context.Background(), "longlonglong2")
	fmt.Println(res, err)

	res, err = demo.DemoBusiness.GetMessage(context.Background())
	fmt.Println(res, err)

	return nil
}

func RedisSetHandler(c *cli.Context) error {
	ctx := c.Context

	cacheKey := c.Args().Get(0)
	cacheValue := c.Args().Get(1)
	expire := cast.ToInt(c.Args().Get(2))

	if cacheKey == "" {
		return fmt.Errorf("cache_key is empty")
	}
	if cacheValue == "" {
		return fmt.Errorf("cache_value is empty")
	}

	result, err := webcache.RedisClient.Set(ctx, cacheKey, cacheValue, time.Second*time.Duration(expire)).Result()
	if err != nil {
		return fmt.Errorf("set cache_key: %s failed, err: %w", cacheKey, err)
	}
	fmt.Printf("cache_key: %s, cache_value: %s, expire: %d, result: %s\n", cacheKey, cacheValue, expire, result)

	return nil
}

func RedisGetHandler(c *cli.Context) error {
	ctx := c.Context

	cacheKey := c.Args().Get(0)

	if cacheKey == "" {
		return fmt.Errorf("cache_key is empty")
	}

	cacheValue, err := webcache.RedisClient.Get(ctx, cacheKey).Result()
	if err == redis.Nil {
    	return fmt.Errorf("cache_key: %s not found", cacheKey)
	} else if err != nil {
		return fmt.Errorf("get cache_key: %s failed, err: %w", cacheKey, err)
	}

	fmt.Println(cacheValue)
	return nil
}
