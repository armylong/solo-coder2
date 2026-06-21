package index

import (
	desktopModel "github.com/armylong/armylong-go/internal/model/desktop"
	userModel "github.com/armylong/armylong-go/internal/model/user"
)

type DesktopOsRequest struct{}

type DesktopAppInfo struct {
	AppId   int64  `json:"app_id"`
	AppName string `json:"app_name"`
	X       int    `json:"x"`
	Y       int    `json:"y"`
}

type DockAppInfo struct {
	AppId     int64  `json:"app_id"`
	AppName   string `json:"app_name"`
	DockIndex int    `json:"dock_index"`
}

type UserAppLayout struct {
	DesktopApps []*DesktopAppInfo `json:"desktop_apps"`
	DockApps    []*DockAppInfo    `json:"dock_apps"`
}

type DesktopOsResponse struct {
	User           *userModel.TbUser     `json:"user"`
	UserPermission int                   `json:"user_permission"`
	Apps           []*desktopModel.TbApp `json:"apps"`
	Layout         *UserAppLayout        `json:"layout"`
}
