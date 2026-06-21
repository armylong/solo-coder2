package index

import (
	"context"

	indexCs "github.com/armylong/armylong-go/internal/cs/index"
	"github.com/armylong/armylong-go/internal/middlewares"
	desktopModel "github.com/armylong/armylong-go/internal/model/desktop"
	userModel "github.com/armylong/armylong-go/internal/model/user"
)

type homeBusiness struct{}

var HomeBusiness = &homeBusiness{}

func (h *homeBusiness) DesktopOs(ctx context.Context, req *indexCs.DesktopOsRequest) (res *indexCs.DesktopOsResponse, err error) {
	uid := middlewares.GetLoginUIDFromContext(ctx)
	if uid == 0 {
		return nil, nil
	}

	user, err := userModel.TbUserModel.GetByUid(uid)
	if err != nil || user == nil {
		return nil, nil
	}

	userPermission := userModel.TbAdminUserModel.GetUserPermission(uid)

	longStoreApp := h.getOrCreateLongStoreApp()

	userHasLongStore := h.checkUserHasLongStore(uid, longStoreApp.AppId)
	if !userHasLongStore {
		h.installLongStoreForUser(uid, longStoreApp.AppId)
	}

	userApps, err := desktopModel.TbUserAppModel.ListByUid(uid)
	if err != nil {
		userApps = []*desktopModel.TbUserApp{}
	}

	installedAppMap := make(map[int64]*desktopModel.TbUserApp)
	for _, ua := range userApps {
		if ua.Status == desktopModel.UserAppStatusInstalled {
			installedAppMap[ua.AppId] = ua
		}
	}

	allApps, err := desktopModel.TbAppModel.ListByPermission(userPermission)
	if err != nil {
		allApps = []*desktopModel.TbApp{}
	}

	var installedApps []*desktopModel.TbApp
	for _, app := range allApps {
		if _, exists := installedAppMap[app.AppId]; exists {
			installedApps = append(installedApps, app)
		}
	}

	if len(installedApps) == 0 {
		installedApps = append(installedApps, longStoreApp)
	}

	layout := h.getUserAppLayout(uid, installedApps, installedAppMap)

	return &indexCs.DesktopOsResponse{
		User:           user,
		UserPermission: userPermission,
		Apps:           installedApps,
		Layout:         layout,
	}, nil
}

func (h *homeBusiness) getOrCreateLongStoreApp() *desktopModel.TbApp {
	allApps, err := desktopModel.TbAppModel.ListByPermission(userModel.UserPermissionNormal)
	if err != nil {
		allApps = []*desktopModel.TbApp{}
	}

	for _, app := range allApps {
		if desktopModel.TbAppModel.IsLongStoreApp(app.AppName, app.Url) {
			return app
		}
	}

	newApp := &desktopModel.TbApp{
		AppName:    desktopModel.LongStoreAppName,
		Desc:       "应用商店，浏览和安装各类应用",
		Icon:       "🏪",
		Url:        desktopModel.LongStoreAppUrl,
		Type:       desktopModel.AppTypeApplication,
		Permission: userModel.UserPermissionNormal,
		Status:     1,
	}
	_, _ = desktopModel.TbAppModel.Create(newApp)

	createdApp, _ := desktopModel.TbAppModel.GetByAppName(desktopModel.LongStoreAppName)
	if createdApp != nil {
		return createdApp
	}

	return newApp
}

func (h *homeBusiness) checkUserHasLongStore(uid, longStoreAppId int64) bool {
	userApp, _ := desktopModel.TbUserAppModel.GetByUidAndAppId(uid, longStoreAppId)
	return userApp != nil && userApp.Status == desktopModel.UserAppStatusInstalled
}

func (h *homeBusiness) installLongStoreForUser(uid, longStoreAppId int64) {
	x, y := desktopModel.TbUserAppModel.FindNextAvailablePosition(uid)
	ext := &desktopModel.UserAppExt{
		Position: desktopModel.UserAppPositionDesktop,
		X:        x,
		Y:        y,
	}
	_ = desktopModel.TbUserAppModel.CreateOrUpdate(
		uid,
		longStoreAppId,
		ext,
		desktopModel.UserAppStatusInstalled,
	)
}

func (h *homeBusiness) getUserAppLayout(uid int64, apps []*desktopModel.TbApp, installedAppMap map[int64]*desktopModel.TbUserApp) *indexCs.UserAppLayout {
	if len(apps) == 0 {
		return &indexCs.UserAppLayout{
			DesktopApps: []*indexCs.DesktopAppInfo{},
			DockApps:    []*indexCs.DockAppInfo{},
		}
	}

	desktopApps := make([]*indexCs.DesktopAppInfo, 0)
	dockApps := make([]*indexCs.DockAppInfo, 0)

	for _, app := range apps {
		ua, exists := installedAppMap[app.AppId]
		if !exists || ua.Ext == nil {
			continue
		}

		if ua.Ext.Position == desktopModel.UserAppPositionDock {
			dockApps = append(dockApps, &indexCs.DockAppInfo{
				AppId:     app.AppId,
				AppName:   app.AppName,
				DockIndex: ua.Ext.DockIndex,
			})
		} else {
			desktopApps = append(desktopApps, &indexCs.DesktopAppInfo{
				AppId:   app.AppId,
				AppName: app.AppName,
				X:       ua.Ext.X,
				Y:       ua.Ext.Y,
			})
		}
	}

	for i := 0; i < len(dockApps); i++ {
		for j := i + 1; j < len(dockApps); j++ {
			if dockApps[i].DockIndex > dockApps[j].DockIndex {
				dockApps[i], dockApps[j] = dockApps[j], dockApps[i]
			}
		}
	}

	return &indexCs.UserAppLayout{
		DesktopApps: desktopApps,
		DockApps:    dockApps,
	}
}
