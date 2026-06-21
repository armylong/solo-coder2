package user

import (
	"context"
	"errors"

	userBiz "github.com/armylong/armylong-go/internal/business/user"
	"github.com/armylong/armylong-go/internal/middlewares"
	userCs "github.com/armylong/armylong-go/internal/cs/user"
	"github.com/armylong/armylong-go/internal/model/user"
	"github.com/gin-gonic/gin"
)

// AuthController 认证相关
type AuthController struct{}

// 注册
func (c *AuthController) ActionRegister(ctx *gin.Context, req *userBiz.RegisterRequest) (*userBiz.LoginResponse, error) {
	return userBiz.Register(req)
}

// 登录
func (c *AuthController) ActionLogin(ctx *gin.Context, req *userBiz.LoginRequest) (*userBiz.LoginResponse, error) {
	return userBiz.Login(req)
}

// 登出
func (c *AuthController) ActionLogout(ctx *gin.Context) error {
	token := middlewares.GetLoginToken(ctx)
	return userBiz.Logout(token)
}

// 踢下线
func (c *AuthController) ActionKickoff(ctx *gin.Context, req *userBiz.KickoffRequest) error {
	return userBiz.Kickoff(req.Uid, req.DeviceType)
}

// UserController 用户信息
type UserController struct{}

// 获取当前用户信息
func (c *UserController) ActionGetUserInfo(ctx context.Context, req *userCs.GetUserInfoRequest) (*userCs.GetUserInfoResponse, error) {
	uid := middlewares.GetLoginUIDFromContext(ctx)
	if uid == 0 {
		return nil, errors.New("请先登录")
	}

	u, err := user.TbUserModel.GetByUid(uid)
	if err != nil || u == nil {
		return nil, errors.New("用户不存在")
	}

	u.ClearPassword()

	return &userCs.GetUserInfoResponse{
		Uid:       u.Uid,
		Account:   u.Account,
		Name:      u.Name,
		Email:     u.Email,
		Phone:     u.Phone,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}, nil
}

// 更新当前用户信息
func (c *UserController) ActionUpdateUserInfo(ctx context.Context, req *userCs.UpdateUserInfoRequest) (*userCs.UpdateUserInfoResponse, error) {
	uid := middlewares.GetLoginUIDFromContext(ctx)
	if uid == 0 {
		return nil, errors.New("请先登录")
	}

	u, err := user.TbUserModel.GetByUid(uid)
	if err != nil || u == nil {
		return nil, errors.New("用户不存在")
	}

	if req.Name != "" {
		u.Name = req.Name
	}
	if req.Email != "" {
		u.Email = req.Email
	}
	if req.Phone != "" {
		u.Phone = req.Phone
	}

	err = user.TbUserModel.Update(u)
	if err != nil {
		return nil, errors.New("更新失败: " + err.Error())
	}

	u.ClearPassword()

	return &userCs.UpdateUserInfoResponse{
		Uid:     u.Uid,
		Account: u.Account,
		Name:    u.Name,
		Email:   u.Email,
		Phone:   u.Phone,
		Status:  u.Status,
	}, nil
}

// 修改密码
func (c *UserController) ActionChangePassword(ctx context.Context, req *userBiz.ChangePasswordRequest) error {
	uid := middlewares.GetLoginUIDFromContext(ctx)
	if uid == 0 {
		return errors.New("请先登录")
	}

	return userBiz.ChangePassword(uid, req)
}

// UserManagementController 用户管理（后台）
type UserManagementController struct{}

// 用户统计
func (c *UserManagementController) ActionStats(ctx context.Context, req *userBiz.StatsRequest) (*userBiz.StatsResponse, error) {
	return userBiz.UserBusiness.Stats(ctx, req)
}

// 用户列表
func (c *UserManagementController) ActionUserList(ctx context.Context, req *userBiz.UserListRequest) (*userBiz.UserListResponse, error) {
	return userBiz.UserBusiness.UserList(ctx, req)
}

// 更新用户状态
func (c *UserManagementController) ActionUpdateStatus(ctx context.Context, req *userBiz.UpdateStatusRequest) error {
	return userBiz.UserBusiness.UpdateStatus(ctx, req)
}

// 踢下线
func (c *UserManagementController) ActionKickoff(ctx context.Context, req *userBiz.KickoffRequest) error {
	return userBiz.UserBusiness.KickoffUser(ctx, req)
}

// 修改密码
func (c *UserManagementController) ActionUpdatePassword(ctx context.Context, req *userBiz.UpdatePasswordRequest) error {
	return userBiz.UserBusiness.UpdatePassword(ctx, req)
}

// 设置管理员
func (c *UserManagementController) ActionUpdateAdmin(ctx context.Context, req *userBiz.UpdateAdminRequest) error {
	return userBiz.UserBusiness.UpdateAdmin(ctx, req)
}
