package desktop

import (
	"strings"
	"time"

	userModel "github.com/armylong/armylong-go/internal/model/user"
	"github.com/armylong/go-library/service/sqlite"
)

type TbApp struct {
	AppId      int64     `json:"app_id" db:"pk"`
	AppName    string    `json:"app_name"`
	Desc       string    `json:"desc"`
	Icon       string    `json:"icon"`
	Url        string    `json:"url"`
	Type       int       `json:"type"`
	Permission int       `json:"permission"`
	Status     int       `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type tbAppModel struct{}

var TbAppModel = &tbAppModel{}

func init() {
	_ = TbAppModel.CreateTable()
	_ = TbAppModel.InitDefaultApps()
}

func (m *tbAppModel) TableName() string {
	return "tb_app"
}

func (m *tbAppModel) CreateTable() error {
	sql := `
	CREATE TABLE IF NOT EXISTS tb_app (
		app_id INTEGER PRIMARY KEY AUTOINCREMENT,
		app_name TEXT NOT NULL UNIQUE,
		desc TEXT,
		icon TEXT,
		url TEXT,
		type INTEGER DEFAULT 1,
		permission INTEGER DEFAULT 0,
		status INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err := sqlite.DB.DB().Exec(sql)
	if err != nil {
		return err
	}
	return sqlite.DB.AutoMigrate(m.TableName(), &TbApp{})
}

const (
	LongStoreAppName = "Long Store"
	LongStoreAppUrl  = "long-store"
)

func defaultApps() []*TbApp {
	return []*TbApp{
		{AppName: LongStoreAppName, Desc: "应用商店，浏览和安装各类应用", Icon: "🏪", Url: LongStoreAppUrl, Type: AppTypeApplication, Permission: userModel.UserPermissionNormal, Status: 1},
	}
}

func (m *tbAppModel) InitDefaultApps() error {
	for _, app := range defaultApps() {
		existing, err := m.GetByAppName(app.AppName)
		if err != nil || existing == nil {
			_, _ = m.Create(app)
		} else {
			if existing.Permission != app.Permission {
				existing.Permission = app.Permission
				_ = m.Update(existing)
			}
		}
	}
	return nil
}

func (m *tbAppModel) Create(app *TbApp) (int64, error) {
	app.CreatedAt = time.Now()
	app.UpdatedAt = time.Now()
	return sqlite.DB.Insert(m.TableName(), app)
}

func (m *tbAppModel) GetByAppId(appId int64) (*TbApp, error) {
	var app TbApp
	app.AppId = appId
	err := sqlite.DB.GetByPkId(m.TableName(), &app)
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (m *tbAppModel) GetByAppName(appName string) (*TbApp, error) {
	var app TbApp
	err := sqlite.DB.FindOne(m.TableName(), &app, "app_name = ?", appName)
	if err != nil {
		return nil, err
	}
	return &app, nil
}

func (m *tbAppModel) ListByPermission(userPermission int) ([]*TbApp, error) {
	var apps []*TbApp
	var err error

	switch {
	case userPermission >= userModel.UserPermissionSuperAdmin:
		err = sqlite.DB.Find(m.TableName(), &apps, "status = 1 ORDER BY type, app_id")
	case userPermission >= userModel.UserPermissionAdmin:
		err = sqlite.DB.Find(m.TableName(), &apps, "status = 1 AND (permission = 0 OR permission = 1) ORDER BY type, app_id")
	default:
		err = sqlite.DB.Find(m.TableName(), &apps, "status = 1 AND permission = 0 ORDER BY type, app_id")
	}

	return apps, err
}

func (m *tbAppModel) Update(app *TbApp) error {
	app.UpdatedAt = time.Now()
	return sqlite.DB.UpdateByPkId(m.TableName(), app)
}

func (m *tbAppModel) Delete(appId int64) error {
	app := &TbApp{AppId: appId}
	return sqlite.DB.DeleteByPkId(m.TableName(), app)
}

func (m *tbAppModel) ListForLongStore(userPermission int) ([]*TbApp, error) {
	var allApps []*TbApp
	var err error

	switch {
	case userPermission >= userModel.UserPermissionSuperAdmin:
		err = sqlite.DB.Find(m.TableName(), &allApps, "status = 1 AND url != ?", LongStoreAppUrl)
	case userPermission >= userModel.UserPermissionAdmin:
		err = sqlite.DB.Find(m.TableName(), &allApps, "status = 1 AND url != ? AND (permission = 0 OR permission = 1)", LongStoreAppUrl)
	default:
		err = sqlite.DB.Find(m.TableName(), &allApps, "status = 1 AND url != ? AND permission = 0", LongStoreAppUrl)
	}

	if err != nil {
		return nil, err
	}

	var superAdminApps, adminApps, userApps []*TbApp
	for _, app := range allApps {
		switch {
		case app.Permission >= userModel.UserPermissionSuperAdmin:
			superAdminApps = append(superAdminApps, app)
		case app.Permission >= userModel.UserPermissionAdmin:
			adminApps = append(adminApps, app)
		default:
			userApps = append(userApps, app)
		}
	}

	result := make([]*TbApp, 0, len(allApps))
	result = append(result, superAdminApps...)
	result = append(result, adminApps...)
	result = append(result, userApps...)

	return result, nil
}

// IsLongStoreApp 判断是否是long-store应用
func (m *tbAppModel) IsLongStoreApp(appName string, url string) bool {
	// 如果名称或url包含定义的字符串，返回true
	if strings.Contains(appName, LongStoreAppName) || strings.Contains(url, LongStoreAppUrl) {
		return true
	}
	return false
}
