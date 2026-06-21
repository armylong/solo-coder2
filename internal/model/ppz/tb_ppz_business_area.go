package ppz

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/armylong/go-library/service/sqlite"
)

// 区域状态
const (
	BusinessAreaStatusDeleted  = 0 // 已删除
	BusinessAreaStatusNormal   = 1 // 正常
	BusinessAreaStatusDisabled = 2 // 已停用
)

// 区域围栏数据
type AreaFence [][][][]float64

func (f AreaFence) Value() (driver.Value, error) {
	return json.Marshal(f)
}

func (f *AreaFence) Scan(value interface{}) error {
	if value == nil {
		*f = nil
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into AreaFence", value)
	}
	return json.Unmarshal(bytes, f)
}

// 运营区域
type TbPpzBusinessArea struct {
	AreaId    int64      `json:"area_id" db:"pk"` // 区域ID
	AreaName  string     `json:"area_name"`       // 区域名称
	AreaFence *AreaFence `json:"area_fence"`      // 围栏数据
	Status    int        `json:"status"`          // 状态
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type tbPpzBusinessAreaModel struct{}

var TbPpzBusinessAreaModel = &tbPpzBusinessAreaModel{}

func init() {
	_ = TbPpzBusinessAreaModel.CreateTable()
}

func (m *tbPpzBusinessAreaModel) TableName() string {
	return "tb_ppz_business_area"
}

// 建表
func (m *tbPpzBusinessAreaModel) CreateTable() error {
	sql := `
	CREATE TABLE IF NOT EXISTS tb_ppz_business_area (
		area_id INTEGER PRIMARY KEY AUTOINCREMENT,
		area_name TEXT NOT NULL DEFAULT '',
		area_fence TEXT,
		status INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err := sqlite.DB.DB().Exec(sql)
	if err != nil {
		return err
	}
	return sqlite.DB.AutoMigrate(m.TableName(), &TbPpzBusinessArea{})
}

// 新增区域
func (m *tbPpzBusinessAreaModel) Create(area *TbPpzBusinessArea) (int64, error) {
	if area.Status == 0 {
		area.Status = BusinessAreaStatusNormal
	}
	return sqlite.DB.Insert(m.TableName(), area)
}

// 按id查询
func (m *tbPpzBusinessAreaModel) GetById(areaId int64) (*TbPpzBusinessArea, error) {
	var area TbPpzBusinessArea
	area.AreaId = areaId
	err := sqlite.DB.GetByPkId(m.TableName(), &area)
	if err != nil {
		return nil, err
	}
	return &area, nil
}

// 更新
func (m *tbPpzBusinessAreaModel) Update(area *TbPpzBusinessArea) error {
	area.UpdatedAt = time.Now()
	return sqlite.DB.UpdateByPkId(m.TableName(), area)
}

// 删除（软删除）
func (m *tbPpzBusinessAreaModel) Delete(areaId int64) error {
	area, err := m.GetById(areaId)
	if err != nil {
		return err
	}
	area.Status = BusinessAreaStatusDeleted
	return m.Update(area)
}

// 停用
func (m *tbPpzBusinessAreaModel) Disable(areaId int64) error {
	area, err := m.GetById(areaId)
	if err != nil {
		return err
	}
	area.Status = BusinessAreaStatusDisabled
	return m.Update(area)
}

// 启用
func (m *tbPpzBusinessAreaModel) Enable(areaId int64) error {
	area, err := m.GetById(areaId)
	if err != nil {
		return err
	}
	area.Status = BusinessAreaStatusNormal
	return m.Update(area)
}

// 分页列表
func (m *tbPpzBusinessAreaModel) List(status int, keyword string, limit, offset int) ([]*TbPpzBusinessArea, error) {
	var areas []*TbPpzBusinessArea
	var err error

	if status > 0 && keyword != "" {
		err = sqlite.DB.Find(m.TableName(), &areas, "status = ? AND area_name LIKE ? ORDER BY updated_at DESC LIMIT ? OFFSET ?", status, "%"+keyword+"%", limit, offset)
	} else if status > 0 {
		err = sqlite.DB.Find(m.TableName(), &areas, "status = ? ORDER BY updated_at DESC LIMIT ? OFFSET ?", status, limit, offset)
	} else if keyword != "" {
		err = sqlite.DB.Find(m.TableName(), &areas, "area_name LIKE ? ORDER BY updated_at DESC LIMIT ? OFFSET ?", "%"+keyword+"%", limit, offset)
	} else {
		err = sqlite.DB.Find(m.TableName(), &areas, "status > 0 ORDER BY updated_at DESC LIMIT ? OFFSET ?", limit, offset)
	}
	return areas, err
}

// 按条件统计
func (m *tbPpzBusinessAreaModel) Count(status int, keyword string) (int64, error) {
	if status > 0 && keyword != "" {
		return sqlite.DB.Count(m.TableName(), "status = ? AND area_name LIKE ?", status, "%"+keyword+"%")
	} else if status > 0 {
		return sqlite.DB.Count(m.TableName(), "status = ?", status)
	} else if keyword != "" {
		return sqlite.DB.Count(m.TableName(), "area_name LIKE ?", "%"+keyword+"%")
	}
	return sqlite.DB.Count(m.TableName(), "status > 0")
}

// 获取启用的区域列表
func (m *tbPpzBusinessAreaModel) ListActive() ([]*TbPpzBusinessArea, error) {
	var areas []*TbPpzBusinessArea
	err := sqlite.DB.Find(m.TableName(), &areas, "status = ? ORDER BY area_name ASC", BusinessAreaStatusNormal)
	return areas, err
}
