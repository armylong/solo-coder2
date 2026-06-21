package user

import (
	"context"
	"errors"

	"github.com/armylong/armylong-go/internal/middlewares"
	"github.com/armylong/armylong-go/internal/model/user"
	"golang.org/x/crypto/bcrypt"
)

type userBusiness struct{}

var UserBusiness = &userBusiness{}

// 检查操作权限（只能操作比自己权限低的用户）
func checkPermission(ctx context.Context, targetUid int64) error {
	loginUser := middlewares.GetLoginUserFromContext(ctx)
	if loginUser == nil {
		return errors.New("请先登录")
	}

	if loginUser.UserPermission < user.UserPermissionAdmin {
		return errors.New("权限不足")
	}

	targetPermission := user.TbAdminUserModel.GetUserPermission(targetUid)
	if targetPermission >= loginUser.UserPermission {
		return errors.New("不能操作大于等于自己权限的用户")
	}

	return nil
}

// 用户统计
func (b *userBusiness) Stats(ctx context.Context, req *StatsRequest) (*StatsResponse, error) {
	totalUsers, err := user.TbUserModel.Count()
	if err != nil {
		return nil, errors.New("获取用户总数失败")
	}

	adminUsers, err := user.TbAdminUserModel.CountAdmins()
	if err != nil {
		return nil, errors.New("获取管理员数量失败")
	}

	return &StatsResponse{
		TotalUsers: totalUsers,
		AdminUsers: adminUsers,
	}, nil
}

// 用户列表（管理后台用）
func (b *userBusiness) UserList(ctx context.Context, req *UserListRequest) (*UserListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	offset := (req.Page - 1) * req.PageSize

	users, err := user.TbUserModel.List(req.PageSize, offset)
	if err != nil {
		return nil, errors.New("获取用户列表失败")
	}

	total, err := user.TbUserModel.Count()
	if err != nil {
		return nil, errors.New("获取用户总数失败")
	}

	adminUids, err := user.TbAdminUserModel.ListAll()
	if err != nil {
		adminUids = []*user.TbAdminUser{}
	}

	adminPermMap := make(map[int64]int)
	for _, admin := range adminUids {
		adminPermMap[admin.Uid] = admin.Permission
	}

	userList := make([]*UserInfo, len(users))
	for i, u := range users {
		u.ClearPassword()
		userList[i] = &UserInfo{
			Uid:        u.Uid,
			Account:    u.Account,
			Name:       u.Name,
			Email:      u.Email,
			Phone:      u.Phone,
			Status:     u.Status,
			Permission: adminPermMap[u.Uid],
		}
	}

	return &UserListResponse{
		Users:    userList,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// 更新用户状态（禁用/启用）
func (b *userBusiness) UpdateStatus(ctx context.Context, req *UpdateStatusRequest) error {
	if req.Uid == 0 {
		return errors.New("用户ID不能为空")
	}

	if err := checkPermission(ctx, req.Uid); err != nil {
		return err
	}

	u, err := user.TbUserModel.GetByUid(req.Uid)
	if err != nil || u == nil {
		return errors.New("用户不存在")
	}

	u.Status = req.Status
	err = user.TbUserModel.Update(u)
	if err != nil {
		return errors.New("更新状态失败: " + err.Error())
	}

	// 禁用时踢下线
	if req.Status == 0 {
		_ = Kickoff(req.Uid, "")
	}

	return nil
}

// 踢下线（管理后台用）
func (b *userBusiness) KickoffUser(ctx context.Context, req *KickoffRequest) error {
	if req.Uid == 0 {
		return errors.New("用户ID不能为空")
	}

	if err := checkPermission(ctx, req.Uid); err != nil {
		return err
	}

	err := Kickoff(req.Uid, req.DeviceType)
	if err != nil {
		return errors.New("踢下线失败: " + err.Error())
	}

	return nil
}

// 修改密码（管理后台用）
func (b *userBusiness) UpdatePassword(ctx context.Context, req *UpdatePasswordRequest) error {
	if req.Uid == 0 {
		return errors.New("用户ID不能为空")
	}

	if err := checkPermission(ctx, req.Uid); err != nil {
		return err
	}

	if req.NewPassword == "" {
		return errors.New("新密码不能为空")
	}

	u, err := user.TbUserModel.GetByUid(req.Uid)
	if err != nil || u == nil {
		return errors.New("用户不存在")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}

	u.Password = string(hashedPassword)
	err = user.TbUserModel.Update(u)
	if err != nil {
		return errors.New("更新密码失败: " + err.Error())
	}

	return nil
}

// 设置/取消管理员
func (b *userBusiness) UpdateAdmin(ctx context.Context, req *UpdateAdminRequest) error {
	if req.Uid == 0 {
		return errors.New("用户ID不能为空")
	}

	if err := checkPermission(ctx, req.Uid); err != nil {
		return err
	}

	_, err := user.TbUserModel.GetByUid(req.Uid)
	if err != nil || err != nil {
		return errors.New("用户不存在")
	}

	err = user.TbAdminUserModel.SetAdmin(req.Uid, req.IsAdmin)
	if err != nil {
		return errors.New("更新管理员状态失败: " + err.Error())
	}

	return nil
}
