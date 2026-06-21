package user

import (
	"errors"

	"github.com/armylong/armylong-go/internal/middlewares"
	"github.com/armylong/armylong-go/internal/model/user"
	"golang.org/x/crypto/bcrypt"
)

// 修改密码（用户自己改）
func ChangePassword(uid int64, req *ChangePasswordRequest) error {
	if req.OldPassword == "" {
		return errors.New("旧密码不能为空")
	}
	if req.NewPassword == "" {
		return errors.New("新密码不能为空")
	}
	if req.ConfirmPassword == "" {
		return errors.New("确认密码不能为空")
	}
	if req.NewPassword != req.ConfirmPassword {
		return errors.New("两次输入的新密码不一致")
	}
	if len(req.NewPassword) < 6 {
		return errors.New("新密码长度至少为6位")
	}

	u, err := user.TbUserModel.GetByUid(uid)
	if err != nil || u == nil {
		return errors.New("用户不存在")
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.OldPassword))
	if err != nil {
		return errors.New("旧密码错误")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}

	u.Password = string(hashedPassword)
	err = user.TbUserModel.Update(u)
	if err != nil {
		return errors.New("更新密码失败")
	}

	// 改完密码踢所有设备下线
	user.TbUserTokenModel.DeleteByUid(u.Uid)
	middlewares.ClearCacheByUID(u.Uid)

	return nil
}
