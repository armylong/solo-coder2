package demo

import (
	"context"

	"github.com/armylong/armylong-go/internal/common/config"
	"github.com/armylong/armylong-go/internal/common/webcache"
)

// 示例业务
type demoBusiness struct{}

var DemoBusiness = &demoBusiness{}

// 设置示例消息到Redis
func (b *demoBusiness) SetMessage(ctx context.Context, message string) (res string, err error) {
	res, err = webcache.RedisClient.Set(ctx, config.DemoMessageCacheKey, message, 0).Result()
	if err != nil {
		return "", err
	}
	return res, nil
}

// 从Redis获取示例消息
func (b *demoBusiness) GetMessage(ctx context.Context) (res string, err error) {
	res, err = webcache.RedisClient.Get(ctx, config.DemoMessageCacheKey).Result()
	if err != nil {
		return "", err
	}
	return res, nil
}
