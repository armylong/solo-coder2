package long_store

import (
	desktopModel "github.com/armylong/armylong-go/internal/model/desktop"
)

// 应用列表-请求
type AppListRequest struct{}

// 应用列表项
type AppListItem struct {
	AppId       int64  `json:"app_id"`       // 应用ID
	AppName     string `json:"app_name"`     // 应用名称
	Desc        string `json:"desc"`         // 描述
	Icon        string `json:"icon"`         // 图标
	Url         string `json:"url"`          // 访问地址
	Type        int    `json:"type"`         // 类型: 1-应用 2-游戏
	Permission  int    `json:"permission"`   // 权限等级
	IsInstalled bool   `json:"is_installed"` // 是否已安装
}

// 应用列表-响应
type AppListResponse struct {
	Applications []*AppListItem `json:"applications"` // 应用列表
	Games        []*AppListItem `json:"games"`        // 游戏列表
}

// 安装应用-请求
type InstallAppRequest struct {
	Uid   int64 `json:"uid" form:"uid"`     // 用户ID
	AppId int64 `json:"app_id" form:"app_id"` // 应用ID
}

// 卸载应用-请求
type UninstallAppRequest struct {
	Uid   int64 `json:"uid" form:"uid"`     // 用户ID
	AppId int64 `json:"app_id" form:"app_id"` // 应用ID
}

// 添加应用-请求
type AddAppRequest struct {
	Uid        int64  `json:"uid" form:"uid"`                 // 操作人ID
	AppName    string `json:"app_name" form:"app_name"`       // 应用名称
	Desc       string `json:"desc" form:"desc"`               // 描述
	Icon       string `json:"icon" form:"icon"`               // 图标
	Url        string `json:"url" form:"url"`                 // 访问地址
	Type       int    `json:"type" form:"type"`               // 类型
	Permission int    `json:"permission" form:"permission"`   // 权限等级
}

// 更新应用-请求
type UpdateAppRequest struct {
	Uid        int64  `json:"uid" form:"uid"`                 // 操作人ID
	AppId      int64  `json:"app_id" form:"app_id"`           // 应用ID
	AppName    string `json:"app_name" form:"app_name"`       // 应用名称
	Desc       string `json:"desc" form:"desc"`               // 描述
	Icon       string `json:"icon" form:"icon"`               // 图标
	Url        string `json:"url" form:"url"`                 // 访问地址
	Type       int    `json:"type" form:"type"`               // 类型
	Permission int    `json:"permission" form:"permission"`   // 权限等级
}

// 删除应用-请求
type DeleteAppRequest struct {
	Uid   int64 `json:"uid" form:"uid"`     // 操作人ID
	AppId int64 `json:"app_id" form:"app_id"` // 应用ID
}

// 桌面应用（含位置信息）
type DesktopApp struct {
	*desktopModel.TbApp
	Ext *desktopModel.UserAppExt `json:"ext"` // 位置信息
}

// 桌面应用列表-响应
type DesktopAppsResponse struct {
	DesktopApps []*DesktopApp `json:"desktop_apps"` // 桌面应用
	DockApps    []*DesktopApp `json:"dock_apps"`    // Dock栏应用
}

// 静态路径-请求
type StaticPathsRequest struct{}

// 静态路径项
type StaticPathItem struct {
	Name       string `json:"name"`       // 项目中文名
	Path       string `json:"path"`       // 目录名
	Desc       string `json:"desc"`       // 描述
	Type       int    `json:"type"`       // 类型
	Permission int    `json:"permission"` // 权限等级
}

// 静态路径-响应
type StaticPathsResponse struct {
	Paths []*StaticPathItem `json:"paths"` // 路径列表
}
