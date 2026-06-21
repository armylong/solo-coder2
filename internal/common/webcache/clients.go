package webcache

import "github.com/armylong/go-library/service/redis"

var RedisWslClient *redis.Client // WSL环境Redis客户端
var RedisClient *redis.Client    // 默认Redis客户端

func init() {
	RedisWslClient = redis.GetClient(`wsl`)
	RedisClient = redis.GetClient(`default`)
}
