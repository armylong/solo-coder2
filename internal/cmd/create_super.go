package cmd

import (
	"fmt"

	userBusiness "github.com/armylong/armylong-go/internal/business/user"
	"github.com/armylong/armylong-go/internal/model/user"

	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/bcrypt"
)

// 创建超级管理员
func CreateSuperHandler(c *cli.Context) error {
	account := c.Args().Get(0)
	password := c.Args().Get(1)
	if account == "" || password == "" {
		return fmt.Errorf("用法: create-super <账号> <密码>")
	}

	count, err := user.TbAdminUserModel.CountSuperAdmins()
	if err != nil {
		return fmt.Errorf("检查超管数量失败: %v", err)
	}

	existing, _ := user.TbUserModel.GetByAccount(account)
	if count > 0 {
		permission := user.TbAdminUserModel.GetUserPermission(existing.Uid)
		if permission == user.UserPermissionSuperAdmin {
			// 修改密码
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				return fmt.Errorf("密码加密失败: %v", err)
			}
			existing.Password = string(hashedPassword)
			err = user.TbUserModel.Update(existing)
			if err != nil {
				return fmt.Errorf("更新密码失败: %v", err)
			}
			userBusiness.Kickoff(existing.Uid, "")
			fmt.Printf("用户 %s(%d) 密码已更新\n", existing.Name, existing.Uid)
			return nil
		} else {
			return fmt.Errorf("已存在超级管理员，不允许重复创建。如需修改，请先取消现有超管权限")
		}
	}

	if existing != nil {
		err := user.TbAdminUserModel.SetPermission(existing.Uid, user.UserPermissionSuperAdmin)
		if err != nil {
			return fmt.Errorf("设置权限失败: %v", err)
		}
		fmt.Printf("用户 %s(%d) 已提升为超级管理员\n", existing.Name, existing.Uid)
		return nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %v", err)
	}

	u := &user.TbUser{
		Account:  account,
		Password: string(hashedPassword),
		Name:     account,
		Status:   1,
	}
	_, err = user.TbUserModel.Create(u)
	if err != nil {
		return fmt.Errorf("创建用户失败: %v", err)
	}

	created, _ := user.TbUserModel.GetByAccount(account)
	err = user.TbAdminUserModel.SetPermission(created.Uid, user.UserPermissionSuperAdmin)
	if err != nil {
		return fmt.Errorf("设置权限失败: %v", err)
	}

	fmt.Printf("超级管理员创建成功: %s(%d)\n", created.Name, created.Uid)
	return nil
}
