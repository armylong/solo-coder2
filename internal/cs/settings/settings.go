package settings

import (
	userModel "github.com/armylong/armylong-go/internal/model/user"
)

// 更新用户设置-请求
type UpdateSettingsRequest struct {
	Uid      int64                    `json:"uid" form:"uid"`             // 用户ID
	Settings *userModel.TbUserSetting `json:"settings" form:"settings"`  // 设置内容
}

// 设置桌面应用-请求
type SetDesktopAppRequest struct {
	Uid   int64 `json:"uid" form:"uid"`       // 用户ID
	AppId int64 `json:"app_id" form:"app_id"` // 应用ID
	X     int   `json:"x" form:"x"`           // 桌面X坐标(0-100)
	Y     int   `json:"y" form:"y"`           // 桌面Y坐标(0-100)
}

// 设置Dock栏应用-请求
type SetDockAppRequest struct {
	Uid       int64 `json:"uid" form:"uid"`               // 用户ID
	AppId     int64 `json:"app_id" form:"app_id"`         // 应用ID
	DockIndex int   `json:"dock_index" form:"dock_index"` // Dock栏位置索引
}

// 移除应用-请求
type RemoveAppRequest struct {
	Uid   int64 `json:"uid" form:"uid"`       // 用户ID
	AppId int64 `json:"app_id" form:"app_id"` // 应用ID
}
