package user

import "time"

// Demo消息
type DemoMessage struct {
	Message string `json:"message"` // 消息内容
}

// 获取用户信息-请求
type GetUserInfoRequest struct{}

// 获取用户信息-响应
type GetUserInfoResponse struct {
	Uid       int64     `json:"uid"`        // 用户ID
	Account   string    `json:"account"`    // 账号
	Name      string    `json:"name"`       // 用户名
	Email     string    `json:"email"`      // 邮箱
	Phone     string    `json:"phone"`      // 手机号
	Status    int       `json:"status"`     // 状态
	CreatedAt time.Time `json:"created_at"` // 创建时间
	UpdatedAt time.Time `json:"updated_at"` // 更新时间
}

// 更新用户信息-请求
type UpdateUserInfoRequest struct {
	Name  string `json:"name" form:"name"`   // 用户名
	Email string `json:"email" form:"email"` // 邮箱
	Phone string `json:"phone" form:"phone"` // 手机号
}

// 更新用户信息-响应
type UpdateUserInfoResponse struct {
	Uid     int64  `json:"uid"`     // 用户ID
	Account string `json:"account"` // 账号
	Name    string `json:"name"`    // 用户名
	Email   string `json:"email"`   // 邮箱
	Phone   string `json:"phone"`   // 手机号
	Status  int    `json:"status"`  // 状态
}
