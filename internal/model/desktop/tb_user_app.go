package desktop

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/armylong/go-library/service/sqlite"
)

type UserAppExt struct {
	Position  string `json:"position"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
	DockIndex int    `json:"dock_index"`
}

func (e UserAppExt) Value() (driver.Value, error) {
	return json.Marshal(e)
}

func (e *UserAppExt) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into UserAppExt", value)
	}
	return json.Unmarshal(bytes, e)
}

type TbUserApp struct {
	Id        int64        `json:"id" db:"pk"`
	Uid       int64        `json:"uid"`
	AppId     int64        `json:"app_id"`
	Ext       *UserAppExt  `json:"ext"`
	Status    int          `json:"status"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

type tbUserAppModel struct{}

var TbUserAppModel = &tbUserAppModel{}

func init() {
	_ = TbUserAppModel.CreateTable()
}

func (m *tbUserAppModel) TableName() string {
	return "tb_user_app"
}

func (m *tbUserAppModel) CreateTable() error {
	sql := `
	CREATE TABLE IF NOT EXISTS tb_user_app (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uid INTEGER NOT NULL,
		app_id INTEGER NOT NULL,
		ext TEXT,
		status INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(uid, app_id)
	)
	`
	_, err := sqlite.DB.DB().Exec(sql)
	if err != nil {
		return err
	}
	return sqlite.DB.AutoMigrate(m.TableName(), &TbUserApp{})
}

func (m *tbUserAppModel) Create(userApp *TbUserApp) (int64, error) {
	userApp.CreatedAt = time.Now()
	userApp.UpdatedAt = time.Now()
	return sqlite.DB.Insert(m.TableName(), userApp)
}

func (m *tbUserAppModel) GetByUidAndAppId(uid, appId int64) (*TbUserApp, error) {
	var userApp TbUserApp
	err := sqlite.DB.FindOne(m.TableName(), &userApp, "uid = ? AND app_id = ?", uid, appId)
	if err != nil {
		return nil, err
	}
	return &userApp, nil
}

// ListByUid 查询用户的所有已安装应用
func (m *tbUserAppModel) ListByUid(uid int64) ([]*TbUserApp, error) {
	var userApps []*TbUserApp
	err := sqlite.DB.Find(m.TableName(), &userApps, "uid = ? AND status = ? ORDER BY id", uid, UserAppStatusInstalled)
	return userApps, err
}

// ListDesktopAppsByUid 查询用户的桌面应用
func (m *tbUserAppModel) ListDesktopAppsByUid(uid int64) ([]*TbUserApp, error) {
	userApps, err := m.ListByUid(uid)
	if err != nil {
		return nil, err
	}

	var desktopApps []*TbUserApp
	for _, app := range userApps {
		if app.Ext != nil && app.Ext.Position == UserAppPositionDesktop {
			desktopApps = append(desktopApps, app)
		}
	}
	return desktopApps, nil
}

// ListDockAppsByUid 查询用户的 Dock 栏应用(按 dock_index 排序)
func (m *tbUserAppModel) ListDockAppsByUid(uid int64) ([]*TbUserApp, error) {
	userApps, err := m.ListByUid(uid)
	if err != nil {
		return nil, err
	}

	var dockApps []*TbUserApp
	for _, app := range userApps {
		if app.Ext != nil && app.Ext.Position == UserAppPositionDock {
			dockApps = append(dockApps, app)
		}
	}

	// 按 dock_index 排序
	for i := 0; i < len(dockApps); i++ {
		for j := i + 1; j < len(dockApps); j++ {
			if dockApps[i].Ext.DockIndex > dockApps[j].Ext.DockIndex {
				dockApps[i], dockApps[j] = dockApps[j], dockApps[i]
			}
		}
	}

	return dockApps, nil
}

// CreateOrUpdate 创建或更新用户应用关联
func (m *tbUserAppModel) CreateOrUpdate(uid, appId int64, ext *UserAppExt, status int) error {
	existing, err := m.GetByUidAndAppId(uid, appId)
	if err != nil || existing == nil {
		userApp := &TbUserApp{
			Uid:    uid,
			AppId:  appId,
			Ext:    ext,
			Status: status,
		}
		_, err = m.Create(userApp)
		return err
	}

	existing.Ext = ext
	existing.Status = status
	return m.Update(existing)
}

func (m *tbUserAppModel) Update(userApp *TbUserApp) error {
	userApp.UpdatedAt = time.Now()
	return sqlite.DB.UpdateByPkId(m.TableName(), userApp)
}

// Delete 软删除用户应用关联(更新 status)
func (m *tbUserAppModel) Delete(uid, appId int64) error {
	existing, err := m.GetByUidAndAppId(uid, appId)
	if err != nil || existing == nil {
		return nil
	}
	existing.Status = UserAppStatusUninstalled
	return m.Update(existing)
}

type Position struct {
	X, Y int
}

// FindNextAvailablePosition 查找下一个可用的桌面位置
// 逻辑: 先收集所有已使用的位置，然后从起始位置开始查找第一个空位
// 优先尝试从 0 开始按顺序查找空位，如果没有空位则追加到末尾
func (m *tbUserAppModel) FindNextAvailablePosition(uid int64) (int, int) {
	desktopApps, err := m.ListDesktopAppsByUid(uid)
	if err != nil {
		desktopApps = []*TbUserApp{}
	}

	usedPositions := make(map[Position]bool)
	for _, app := range desktopApps {
		if app.Ext != nil {
			usedPositions[Position{app.Ext.X, app.Ext.Y}] = true
		}
	}

	for i := 0; i < 1000; i++ {
		col := i % DesktopCols
		row := i / DesktopCols
		x := DesktopStartX + col*DesktopGapX
		y := DesktopStartY + row*DesktopGapY
		pos := Position{x, y}
		if !usedPositions[pos] {
			return x, y
		}
	}

	return DesktopStartX, DesktopStartY
}
