package feishu

import (
	"time"

	"github.com/armylong/go-library/service/sqlite"
)

const (
	FeishuBitableConfigStatusDeleted = 0
	FeishuBitableConfigStatusNormal  = 1
)

// TbFeishuBitableConfig 飞书多维表格本地配置
// 用于保存后台接入的表格 token 与别名映射关系
type TbFeishuBitableConfig struct {
	Id         int64     `json:"id" db:"pk"`
	AppToken   string    `json:"app_token"`
	Alias      string    `json:"alias"`
	Status     int       `json:"status"`
	CreatedUid int64     `json:"created_uid"`
	UpdatedUid int64     `json:"updated_uid"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type tbFeishuBitableConfigModel struct{}

var TbFeishuBitableConfigModel = &tbFeishuBitableConfigModel{}

func init() {
	_ = TbFeishuBitableConfigModel.CreateTable()
}

func (m *tbFeishuBitableConfigModel) TableName() string {
	return "tb_feishu_bitable_config"
}

// 建表
func (m *tbFeishuBitableConfigModel) CreateTable() error {
	sql := `
	CREATE TABLE IF NOT EXISTS tb_feishu_bitable_config (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		app_token TEXT NOT NULL UNIQUE,
		alias TEXT NOT NULL DEFAULT '',
		status INTEGER DEFAULT 1,
		created_uid INTEGER DEFAULT 0,
		updated_uid INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err := sqlite.DB.DB().Exec(sql)
	if err != nil {
		return err
	}
	return sqlite.DB.AutoMigrate(m.TableName(), &TbFeishuBitableConfig{})
}

// 新增
func (m *tbFeishuBitableConfigModel) Create(data *TbFeishuBitableConfig) (int64, error) {
	if data.Status == 0 {
		data.Status = FeishuBitableConfigStatusNormal
	}
	return sqlite.DB.Insert(m.TableName(), data)
}

// 按ID查询
func (m *tbFeishuBitableConfigModel) GetById(id int64) (*TbFeishuBitableConfig, error) {
	var data TbFeishuBitableConfig
	data.Id = id
	err := sqlite.DB.GetByPkId(m.TableName(), &data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// 按AppToken查询
func (m *tbFeishuBitableConfigModel) GetByAppToken(appToken string) (*TbFeishuBitableConfig, error) {
	var data TbFeishuBitableConfig
	err := sqlite.DB.FindOne(m.TableName(), &data, "app_token = ?", appToken)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// 更新
func (m *tbFeishuBitableConfigModel) Update(data *TbFeishuBitableConfig) error {
	data.UpdatedAt = time.Now()
	return sqlite.DB.UpdateByPkId(m.TableName(), data)
}

// 删除（软删除）
func (m *tbFeishuBitableConfigModel) Delete(id, updatedUid int64) error {
	data, err := m.GetById(id)
	if err != nil {
		return err
	}
	data.Status = FeishuBitableConfigStatusDeleted
	data.UpdatedUid = updatedUid
	return m.Update(data)
}

// 清空全部配置（软删除）
func (m *tbFeishuBitableConfigModel) DeleteAll() error {
	_, err := sqlite.DB.DB().Exec(
		"UPDATE "+m.TableName()+" SET status = ?, updated_uid = 0, updated_at = CURRENT_TIMESTAMP WHERE status != ?",
		FeishuBitableConfigStatusDeleted,
		FeishuBitableConfigStatusDeleted,
	)
	return err
}

// 分页列表
func (m *tbFeishuBitableConfigModel) List(status int, keyword string, limit, offset int) ([]*TbFeishuBitableConfig, error) {
	var list []*TbFeishuBitableConfig
	var err error

	if status > 0 && keyword != "" {
		err = sqlite.DB.Find(m.TableName(), &list, "status = ? AND (alias LIKE ? OR app_token LIKE ?) ORDER BY updated_at DESC LIMIT ? OFFSET ?", status, "%"+keyword+"%", "%"+keyword+"%", limit, offset)
	} else if status > 0 {
		err = sqlite.DB.Find(m.TableName(), &list, "status = ? ORDER BY updated_at DESC LIMIT ? OFFSET ?", status, limit, offset)
	} else if keyword != "" {
		err = sqlite.DB.Find(m.TableName(), &list, "status > 0 AND (alias LIKE ? OR app_token LIKE ?) ORDER BY updated_at DESC LIMIT ? OFFSET ?", "%"+keyword+"%", "%"+keyword+"%", limit, offset)
	} else {
		err = sqlite.DB.Find(m.TableName(), &list, "status > 0 ORDER BY updated_at DESC LIMIT ? OFFSET ?", limit, offset)
	}
	return list, err
}

// 获取启用配置列表
func (m *tbFeishuBitableConfigModel) ListActive() ([]*TbFeishuBitableConfig, error) {
	var list []*TbFeishuBitableConfig
	err := sqlite.DB.Find(m.TableName(), &list, "status = ? ORDER BY updated_at DESC", FeishuBitableConfigStatusNormal)
	return list, err
}

// 按条件统计
func (m *tbFeishuBitableConfigModel) Count(status int, keyword string) (int64, error) {
	if status > 0 && keyword != "" {
		return sqlite.DB.Count(m.TableName(), "status = ? AND (alias LIKE ? OR app_token LIKE ?)", status, "%"+keyword+"%", "%"+keyword+"%")
	} else if status > 0 {
		return sqlite.DB.Count(m.TableName(), "status = ?", status)
	} else if keyword != "" {
		return sqlite.DB.Count(m.TableName(), "status > 0 AND (alias LIKE ? OR app_token LIKE ?)", "%"+keyword+"%", "%"+keyword+"%")
	}
	return sqlite.DB.Count(m.TableName(), "status > 0")
}
