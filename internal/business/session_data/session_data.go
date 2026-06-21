package session_data

import (
	"context"
	"fmt"
	"sync"

	ppzModel "github.com/armylong/armylong-go/internal/model/ppz"
	userModel "github.com/armylong/armylong-go/internal/model/user"
)

type sessionDataBusiness struct{}

var SessionDataBusiness = &sessionDataBusiness{}

// 数据拉取函数类型
type DataFetcher func(ctx context.Context, uid int64) (interface{}, error)

// 数据拉取器注册表
var dataFetchers = map[string]DataFetcher{
	"user":     fetchUserData,
	"ppz_user": fetchPpzUserData,
}

// 按key列表并发拉取会话数据
func (b *sessionDataBusiness) GetSessionData(ctx context.Context, uid int64, keys []string) (map[string]interface{}, error) {
	if uid == 0 {
		return nil, fmt.Errorf("请先登录")
	}

	result := make(map[string]interface{}, len(keys))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, key := range keys {
		fetcher, ok := dataFetchers[key]
		if !ok {
			continue
		}

		wg.Add(1)
		go func(k string, f DataFetcher) {
			defer wg.Done()
			data, err := f(ctx, uid)
			if err != nil {
				return
			}
			mu.Lock()
			result[k] = data
			mu.Unlock()
		}(key, fetcher)
	}

	wg.Wait()
	return result, nil
}

// 拉取用户基本信息
func fetchUserData(ctx context.Context, uid int64) (interface{}, error) {
	u, err := userModel.TbUserModel.GetByUid(uid)
	if err != nil || u == nil {
		return nil, fmt.Errorf("用户不存在")
	}
	u.ClearPassword()
	return u, nil
}

// 拉取拼拼坐用户信息
func fetchPpzUserData(ctx context.Context, uid int64) (interface{}, error) {
	ppzUser, err := ppzModel.TbPpzUserModel.GetOrCreateByUid(uid)
	if err != nil {
		return nil, fmt.Errorf("获取拼拼坐用户信息失败: %w", err)
	}
	return ppzUser, nil
}
