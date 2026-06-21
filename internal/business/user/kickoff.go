package user

import (
	"errors"

	"github.com/armylong/armylong-go/internal/middlewares"
	"github.com/armylong/armylong-go/internal/model/user"
)

// 踢下线（deviceType为空则踢所有设备）
func Kickoff(uid int64, deviceType string) error {
	if uid == 0 {
		return errors.New("缺少用户ID")
	}

	var tokens []*user.TbUserToken
	var err error

	if deviceType != "" {
		tokens, err = user.TbUserTokenModel.ListByUid(uid)
		if err != nil {
			return errors.New("查询Token失败")
		}
		var filtered []*user.TbUserToken
		for _, t := range tokens {
			if t.DeviceType == deviceType {
				filtered = append(filtered, t)
			}
		}
		tokens = filtered

		user.TbUserTokenModel.DeleteByUidAndDeviceType(uid, deviceType)
	} else {
		tokens, err = user.TbUserTokenModel.ListByUid(uid)
		if err != nil {
			return errors.New("查询Token失败")
		}

		user.TbUserTokenModel.DeleteByUid(uid)
	}

	// 清除缓存
	tokenStrings := make([]string, len(tokens))
	for i, t := range tokens {
		tokenStrings[i] = t.Token
	}
	middlewares.DeleteCacheByTokens(tokenStrings)

	return nil
}
