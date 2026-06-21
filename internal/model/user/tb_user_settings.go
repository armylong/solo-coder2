package user

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/armylong/go-library/service/sqlite"
)

// 用户设置
type TbUserSettings struct {
	Uid       int64          `json:"uid" db:"pk"` // 用户ID
	Settings  *TbUserSetting `json:"settings"`    // 设置内容(JSON)
	UpdatedAt time.Time      `json:"updated_at"`
}

// 用户设置内容
type TbUserSetting struct {
	Desktop TbUserSettingAppList `json:"desktop"` // 桌面应用
	Dock    TbUserSettingAppList `json:"dock"`    // Dock栏应用
}

func (s TbUserSetting) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *TbUserSetting) Scan(value interface{}) error {
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
		return fmt.Errorf("cannot scan %T into TbUserSetting", value)
	}
	return json.Unmarshal(bytes, s)
}

// 应用列表
type TbUserSettingAppList struct {
	AppList []TbUserSettingApp `json:"app_list"` // 应用列表
}

// 单个应用
type TbUserSettingApp struct {
	AppId   string `json:"app_id"`             // 应用ID
	AppName string `json:"app_name,omitempty"` // 应用名
	Desc    string `json:"desc,omitempty"`     // 描述
	X       int    `json:"x"`                  // X坐标
	Y       int    `json:"y"`                  // Y坐标
}

type tbUserSettingsModel struct{}

var TbUserSettingsModel = &tbUserSettingsModel{}

func init() {
	_ = TbUserSettingsModel.CreateTable()
}

func (m *tbUserSettingsModel) TableName() string {
	return "tb_user_settings"
}

// 建表
func (m *tbUserSettingsModel) CreateTable() error {
	sql := `
	CREATE TABLE IF NOT EXISTS tb_user_settings (
		uid INTEGER PRIMARY KEY,
		settings TEXT,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err := sqlite.DB.DB().Exec(sql)
	if err != nil {
		return err
	}
	return sqlite.DB.AutoMigrate(m.TableName(), &TbUserSettings{})
}

// 按用户查设置
func (m *tbUserSettingsModel) GetByUid(uid int64) (*TbUserSettings, error) {
	var settings TbUserSettings
	settings.Uid = uid
	err := sqlite.DB.GetByPkId(m.TableName(), &settings)
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

// 创建或更新设置
func (m *tbUserSettingsModel) CreateOrUpdate(uid int64, settings TbUserSetting) error {
	data := &TbUserSettings{
		Uid:       uid,
		Settings:  &settings,
		UpdatedAt: time.Now(),
	}
	return sqlite.DB.Upsert(m.TableName(), data, "uid")
}

// 删除
func (m *tbUserSettingsModel) Delete(uid int64) error {
	settings := &TbUserSettings{Uid: uid}
	return sqlite.DB.DeleteByPkId(m.TableName(), settings)
}
