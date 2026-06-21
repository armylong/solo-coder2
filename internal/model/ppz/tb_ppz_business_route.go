package ppz

import (
	"time"

	"github.com/armylong/go-library/service/sqlite"
)

// 路线状态
const (
	BusinessRouteStatusDeleted  = 0 // 已删除
	BusinessRouteStatusNormal   = 1 // 正常
	BusinessRouteStatusDisabled = 2 // 已停用
)

// 运营路线
type TbPpzBusinessRoute struct {
	RouteId   int64     `json:"route_id" db:"pk"` // 路线ID
	RouteName string    `json:"route_name"`       // 路线名称
	AAreaId   int64     `json:"a_area_id"`        // A端区域ID
	BAreaId   int64     `json:"b_area_id"`        // B端区域ID
	Status    int       `json:"status"`           // 状态
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type tbPpzBusinessRouteModel struct{}

var TbPpzBusinessRouteModel = &tbPpzBusinessRouteModel{}

func init() {
	_ = TbPpzBusinessRouteModel.CreateTable()
}

func (m *tbPpzBusinessRouteModel) TableName() string {
	return "tb_ppz_business_route"
}

// 建表
func (m *tbPpzBusinessRouteModel) CreateTable() error {
	sql := `
	CREATE TABLE IF NOT EXISTS tb_ppz_business_route (
		route_id INTEGER PRIMARY KEY AUTOINCREMENT,
		route_name TEXT NOT NULL DEFAULT '',
		a_area_id INTEGER NOT NULL DEFAULT 0,
		b_area_id INTEGER NOT NULL DEFAULT 0,
		status INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err := sqlite.DB.DB().Exec(sql)
	if err != nil {
		return err
	}
	return sqlite.DB.AutoMigrate(m.TableName(), &TbPpzBusinessRoute{})
}

// 新增路线
func (m *tbPpzBusinessRouteModel) Create(route *TbPpzBusinessRoute) (int64, error) {
	if route.Status == 0 {
		route.Status = BusinessRouteStatusNormal
	}
	return sqlite.DB.Insert(m.TableName(), route)
}

// 按id查询
func (m *tbPpzBusinessRouteModel) GetById(routeId int64) (*TbPpzBusinessRoute, error) {
	var route TbPpzBusinessRoute
	route.RouteId = routeId
	err := sqlite.DB.GetByPkId(m.TableName(), &route)
	if err != nil {
		return nil, err
	}
	return &route, nil
}

// 更新
func (m *tbPpzBusinessRouteModel) Update(route *TbPpzBusinessRoute) error {
	route.UpdatedAt = time.Now()
	return sqlite.DB.UpdateByPkId(m.TableName(), route)
}

// 删除（软删除）
func (m *tbPpzBusinessRouteModel) Delete(routeId int64) error {
	route, err := m.GetById(routeId)
	if err != nil {
		return err
	}
	route.Status = BusinessRouteStatusDeleted
	return m.Update(route)
}

// 停用
func (m *tbPpzBusinessRouteModel) Disable(routeId int64) error {
	route, err := m.GetById(routeId)
	if err != nil {
		return err
	}
	route.Status = BusinessRouteStatusDisabled
	return m.Update(route)
}

// 启用
func (m *tbPpzBusinessRouteModel) Enable(routeId int64) error {
	route, err := m.GetById(routeId)
	if err != nil {
		return err
	}
	route.Status = BusinessRouteStatusNormal
	return m.Update(route)
}

// 分页列表
func (m *tbPpzBusinessRouteModel) List(status int, keyword string, limit, offset int) ([]*TbPpzBusinessRoute, error) {
	var routes []*TbPpzBusinessRoute
	var err error

	if status > 0 && keyword != "" {
		err = sqlite.DB.Find(m.TableName(), &routes, "status = ? AND route_name LIKE ? ORDER BY updated_at DESC LIMIT ? OFFSET ?", status, "%"+keyword+"%", limit, offset)
	} else if status > 0 {
		err = sqlite.DB.Find(m.TableName(), &routes, "status = ? ORDER BY updated_at DESC LIMIT ? OFFSET ?", status, limit, offset)
	} else if keyword != "" {
		err = sqlite.DB.Find(m.TableName(), &routes, "route_name LIKE ? ORDER BY updated_at DESC LIMIT ? OFFSET ?", "%"+keyword+"%", limit, offset)
	} else {
		err = sqlite.DB.Find(m.TableName(), &routes, "status > 0 ORDER BY updated_at DESC LIMIT ? OFFSET ?", limit, offset)
	}
	return routes, err
}

// 按条件统计
func (m *tbPpzBusinessRouteModel) Count(status int, keyword string) (int64, error) {
	if status > 0 && keyword != "" {
		return sqlite.DB.Count(m.TableName(), "status = ? AND route_name LIKE ?", status, "%"+keyword+"%")
	} else if status > 0 {
		return sqlite.DB.Count(m.TableName(), "status = ?", status)
	} else if keyword != "" {
		return sqlite.DB.Count(m.TableName(), "route_name LIKE ?", "%"+keyword+"%")
	}
	return sqlite.DB.Count(m.TableName(), "status > 0")
}

// 按区域统计关联路线数
func (m *tbPpzBusinessRouteModel) CountByAreaId(areaId int64) (int64, error) {
	return sqlite.DB.Count(m.TableName(), "status > 0 AND (a_area_id = ? OR b_area_id = ?)", areaId, areaId)
}
