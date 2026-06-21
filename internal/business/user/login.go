package user

import (
	"errors"
	"time"

	"github.com/armylong/armylong-go/internal/middlewares"
	"github.com/armylong/armylong-go/internal/model/user"
	"golang.org/x/crypto/bcrypt"
)

// 登录
func Login(req *LoginRequest) (*LoginResponse, error) {
	if req.Account == "" || req.Password == "" {
		return nil, errors.New("账号和密码不能为空")
	}

	deviceType := req.DeviceType
	if deviceType == "" {
		deviceType = "pc"
	}

	u, err := user.TbUserModel.GetByAccount(req.Account)
	if err != nil || u == nil {
		return nil, errors.New("账号不存在")
	}

	if u.Status != 1 {
		return nil, errors.New("账号已被禁用")
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password))
	if err != nil {
		return nil, errors.New("密码错误")
	}

	permission := user.TbAdminUserModel.GetUserPermission(u.Uid)
	token, expireAt, err := middlewares.GenerateToken(u.Uid, u.Name, deviceType, permission, 7*24*time.Hour)
	if err != nil {
		return nil, errors.New("生成Token失败")
	}

	// 同设备踢掉旧Token
	user.TbUserTokenModel.DeleteByUidAndDeviceType(u.Uid, deviceType)

	tokenRecord := &user.TbUserToken{
		Uid:        u.Uid,
		Token:      token,
		DeviceType: deviceType,
		ExpireAt:   expireAt,
	}
	user.TbUserTokenModel.Create(tokenRecord)

	middlewares.SetCache(token, &middlewares.LoginUserInfo{
		TbUser:         u,
		UserPermission: permission,
	})

	return &LoginResponse{
		Token: token,
		User:  u,
	}, nil
}

// 登出
func Logout(token string) error {
	if token != "" {
		middlewares.DeleteCache(token)
		user.TbUserTokenModel.DeleteByToken(token)
	}
	return nil
}
