package user

import (
	"errors"
	"time"

	"github.com/armylong/armylong-go/internal/middlewares"
	"github.com/armylong/armylong-go/internal/model/user"
	"golang.org/x/crypto/bcrypt"
)

// 注册
func Register(req *RegisterRequest) (*LoginResponse, error) {
	if req.Account == "" || req.Password == "" {
		return nil, errors.New("账号和密码不能为空")
	}

	existingUser, _ := user.TbUserModel.GetByAccount(req.Account)
	if existingUser != nil {
		return nil, errors.New("账号已存在")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	u := &user.TbUser{
		Account:  req.Account,
		Password: string(hashedPassword),
		Name:     req.Name,
		Email:    req.Email,
		Phone:    req.Phone,
		Status:   1,
	}

	_, err = user.TbUserModel.Create(u)
	if err != nil {
		return nil, errors.New("创建用户失败: " + err.Error())
	}

	createdUser, _ := user.TbUserModel.GetByAccount(req.Account)

	permission := user.TbAdminUserModel.GetUserPermission(createdUser.Uid)
	token, expireAt, err := middlewares.GenerateToken(createdUser.Uid, createdUser.Name, "pc", permission, 7*24*time.Hour)
	if err != nil {
		return nil, errors.New("生成Token失败")
	}

	tokenRecord := &user.TbUserToken{
		Uid:        createdUser.Uid,
		Token:      token,
		DeviceType: "pc",
		ExpireAt:   expireAt,
	}
	user.TbUserTokenModel.Create(tokenRecord)

	middlewares.SetCache(token, &middlewares.LoginUserInfo{
		TbUser:         createdUser,
		UserPermission: permission,
	})

	return &LoginResponse{
		Token: token,
		User:  createdUser,
	}, nil
}
