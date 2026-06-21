package internal

import (
	"github.com/armylong/armylong-go/internal/common/config"
)

// 注册环境变量同步
// 服务启动时从Redis同步配置到环境变量
func RegisterEnv() {
	config.SyncEnvFromRedis()
}
