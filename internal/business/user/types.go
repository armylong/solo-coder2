package user

import (
	"github.com/armylong/armylong-go/internal/model/user"
)

// 登录-请求
type LoginRequest struct {
	Account    string `json:"account" form:"account"`         // 账号
	Password   string `json:"password" form:"password"`       // 密码
	DeviceType string `json:"device_type" form:"device_type"` // 设备类型
}

// 注册-请求
type RegisterRequest struct {
	Account  string `json:"account" form:"account"`   // 账号
	Password string `json:"password" form:"password"` // 密码
	Name     string `json:"name" form:"name"`         // 用户名
	Email    string `json:"email" form:"email"`       // 邮箱
	Phone    string `json:"phone" form:"phone"`       // 手机号
}

// 登录-响应
type LoginResponse struct {
	Token string       `json:"token"` // JWT Token
	User  *user.TbUser `json:"user"`  // 用户信息
}

// 踢下线-请求
type KickoffRequest struct {
	Uid        int64  `json:"uid" form:"uid"`                   // 用户ID
	DeviceType string `json:"device_type" form:"device_type"` // 设备类型(空=全部)
}

// 修改密码-请求
type ChangePasswordRequest struct {
	OldPassword     string `json:"old_password" form:"old_password"`           // 旧密码
	NewPassword     string `json:"new_password" form:"new_password"`           // 新密码
	ConfirmPassword string `json:"confirm_password" form:"confirm_password"` // 确认密码
}

// 用户统计-请求
type StatsRequest struct{}

// 用户统计-响应
type StatsResponse struct {
	TotalUsers int64 `json:"total_users"` // 总用户数
	AdminUsers int64 `json:"admin_users"` // 管理员数
}

// 用户列表-请求
type UserListRequest struct {
	Page     int `json:"page" form:"page"`           // 页码
	PageSize int `json:"page_size" form:"page_size"` // 每页数量
}

// 用户信息(管理列表用)
type UserInfo struct {
	Uid        int64  `json:"uid"`        // 用户ID
	Account    string `json:"account"`    // 账号
	Name       string `json:"name"`       // 用户名
	Email      string `json:"email"`      // 邮箱
	Phone      string `json:"phone"`      // 手机号
	Status     int    `json:"status"`     // 状态
	Permission int    `json:"permission"` // 权限等级
}

// 用户列表-响应
type UserListResponse struct {
	Users    []*UserInfo `json:"users"`     // 用户列表
	Total    int64       `json:"total"`     // 总数
	Page     int         `json:"page"`      // 当前页
	PageSize int         `json:"page_size"` // 每页数量
}

// 更新状态-请求
type UpdateStatusRequest struct {
	Uid    int64 `json:"uid" form:"uid"`       // 用户ID
	Status int   `json:"status" form:"status"` // 状态
}

// 修改密码(管理员)-请求
type UpdatePasswordRequest struct {
	Uid         int64  `json:"uid" form:"uid"`                       // 用户ID
	NewPassword string `json:"new_password" form:"new_password"` // 新密码
}

// 设置管理员-请求
type UpdateAdminRequest struct {
	Uid     int64 `json:"uid" form:"uid"`             // 用户ID
	IsAdmin bool  `json:"is_admin" form:"is_admin"` // 是否管理员
}
