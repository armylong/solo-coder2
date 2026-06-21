package feishu

import (
	"context"
	"errors"
	"os"

	"github.com/armylong/armylong-go/internal/common/config"
	libraryFeishu "github.com/armylong/go-library/service/feishu"
	"github.com/armylong/go-library/service/redis"
)

type fsConfigBusiness struct{}

var FsConfigBusiness = &fsConfigBusiness{}

const (
	RedisKeyAppId     = "FEISHU_ROBOT_APP_ID"
	RedisKeyAppSecret = "FEISHU_ROBOT_APP_SECRET"
)

func (b *fsConfigBusiness) GetConfig() (appId string, appSecretMasked string) {
	appId, _ = config.GetEnvConfig(RedisKeyAppId)
	appSecretMasked = config.GetEnvConfigMasked(RedisKeyAppSecret)
	return appId, appSecretMasked
}

func (b *fsConfigBusiness) SaveConfig(appId, appSecret string) error {
	if appId == "" {
		return errors.New("App ID不能为空")
	}
	if appSecret == "" {
		return errors.New("App Secret不能为空")
	}

	// 保存到Redis（通用环境变量同步）
	if err := config.SaveEnvConfig(RedisKeyAppId, appId, ""); err != nil {
		return err
	}
	if err := config.SaveEnvConfig(RedisKeyAppSecret, appSecret, "aes"); err != nil {
		return err
	}

	// 立即同步到环境变量
	os.Setenv(RedisKeyAppId, appId)
	os.Setenv(RedisKeyAppSecret, appSecret)

	libraryFeishu.ResetFsConfig()
	libraryFeishu.ResetFeishuSDK()

	return nil
}

func (b *fsConfigBusiness) ValidateConfig() (bool, string) {
	// 重新从Redis同步环境变量
	config.SyncEnvFromRedis()
	libraryFeishu.ResetFsConfig()
	return libraryFeishu.ValidateAppConfig()
}

func (b *fsConfigBusiness) HasUserAccessToken() bool {
	redisClient := redis.GetClient("default")
	ctx := context.Background()
	token, err := redisClient.Get(ctx, libraryFeishu.UserAccessTokenCacheKey).Result()
	return err == nil && token != ""
}

func (b *fsConfigBusiness) DeleteConfig() error {
	// 删除Redis中的配置
	if err := config.DeleteEnvConfig(RedisKeyAppId, RedisKeyAppSecret); err != nil {
		return err
	}

	// 删除用户token
	redisClient := redis.GetClient("default")
	ctx := context.Background()
	redisClient.Del(ctx, libraryFeishu.UserAccessTokenCacheKey)
	redisClient.Del(ctx, libraryFeishu.UserAccessTokenRefreshCacheKey)

	// 清除环境变量
	os.Unsetenv(RedisKeyAppId)
	os.Unsetenv(RedisKeyAppSecret)

	libraryFeishu.ResetFsConfig()
	libraryFeishu.ResetFeishuSDK()

	return nil
}
