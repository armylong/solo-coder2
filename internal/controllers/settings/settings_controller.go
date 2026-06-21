package settings

import (
	"errors"

	settingsBusiness "github.com/armylong/armylong-go/internal/business/settings"
	"github.com/armylong/armylong-go/internal/cs/settings"
	"github.com/gin-gonic/gin"
)

// 设置控制器
type SettingsController struct{}

// 已弃用，请使用set-desktop-app和set-dock-app
func (c *SettingsController) ActionUpdate(ctx *gin.Context, req *settings.UpdateSettingsRequest) error {
	return errors.New("该接口已弃用，请使用新接口: /settings/set-desktop-app, /settings/set-dock-app")
}

// 设置桌面应用位置
func (c *SettingsController) ActionSetDesktopApp(ctx *gin.Context, req *settings.SetDesktopAppRequest) error {
	if req.Uid <= 0 {
		return errors.New("用户ID不能为空")
	}
	if req.AppId <= 0 {
		return errors.New("应用ID不能为空")
	}
	if req.X < 0 || req.X > 100 || req.Y < 0 || req.Y > 100 {
		return errors.New("坐标必须在0-100之间")
	}
	return settingsBusiness.SettingsBusiness.SetDesktopApp(req)
}

// 设置Dock栏应用位置
func (c *SettingsController) ActionSetDockApp(ctx *gin.Context, req *settings.SetDockAppRequest) error {
	if req.Uid <= 0 {
		return errors.New("用户ID不能为空")
	}
	if req.AppId <= 0 {
		return errors.New("应用ID不能为空")
	}
	if req.DockIndex < 0 {
		return errors.New("Dock位置索引不能为负数")
	}
	return settingsBusiness.SettingsBusiness.SetDockApp(req)
}

// 移除应用
func (c *SettingsController) ActionRemoveApp(ctx *gin.Context, req *settings.RemoveAppRequest) error {
	if req.Uid <= 0 {
		return errors.New("用户ID不能为空")
	}
	if req.AppId <= 0 {
		return errors.New("应用ID不能为空")
	}
	return settingsBusiness.SettingsBusiness.RemoveApp(req)
}
