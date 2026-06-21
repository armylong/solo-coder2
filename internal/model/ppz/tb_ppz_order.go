package ppz

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/armylong/go-library/service/sqlite"
)

// 订单状态
const (
	OrderStatusCancelled   = 0 // 已取消
	OrderStatusMatching    = 1 // 匹配中
	OrderStatusAccepted    = 2 // 已接单
	OrderStatusConfirmed   = 3 // 已确认
	OrderStatusMatchFailed = 4 // 匹配失败
)

// 时间类型
const (
	TimeTypeDepart = 1 // 出发时间
	TimeTypeArrive = 2 // 到达时间
)

// 订单高德地址数据
type OrderGaodeData struct {
	FormattedAddress string  `json:"formatted_address"` // 详细地址
	Lng              float64 `json:"lng"`               // 经度
	Lat              float64 `json:"lat"`               // 纬度
	Province         string  `json:"province"`          // 省
	City             string  `json:"city"`              // 市
	District         string  `json:"district"`          // 区
	Township         string  `json:"township"`          // 乡镇
	Adcode           string  `json:"adcode"`            // 区域编码
	AoiName          string  `json:"aoi_name"`          // 兴趣区名称
}

func (d OrderGaodeData) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *OrderGaodeData) Scan(value interface{}) error {
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
		return fmt.Errorf("cannot scan %T into OrderGaodeData", value)
	}
	return json.Unmarshal(bytes, d)
}

// 订单
type TbPpzOrder struct {
	OrderId        int64           `json:"order_id" db:"pk"` // 订单ID
	Uid            int64           `json:"uid"`              // 下单用户ID
	TripId         int64           `json:"trip_id"`          // 关联行程ID
	StartGaodeData *OrderGaodeData `json:"start_gaode_data"` // 出发地高德数据
	DestGaodeData  *OrderGaodeData `json:"dest_gaode_data"`  // 目的地高德数据
	DepartTime     string          `json:"depart_time"`      // 出发时间
	TimeType       int             `json:"time_type"`        // 时间类型: 1-出发 2-到达
	TimeFlex       int             `json:"time_flex"`        // 时间灵活度(分钟)
	PassengerCount int             `json:"passenger_count"`  // 乘客数
	IsCharter      int             `json:"is_charter"`       // 是否包车: 0-否 1-是
	OrderStatus    int             `json:"order_status"`     // 订单状态
	CancelReason   string          `json:"cancel_reason"`    // 取消原因
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

type tbPpzOrderModel struct{}

var TbPpzOrderModel = &tbPpzOrderModel{}

func init() {
	_ = TbPpzOrderModel.CreateTable()
}

func (m *tbPpzOrderModel) TableName() string {
	return "tb_ppz_order"
}

// 建表
func (m *tbPpzOrderModel) CreateTable() error {
	sql := `
	CREATE TABLE IF NOT EXISTS tb_ppz_order (
		order_id INTEGER PRIMARY KEY AUTOINCREMENT,
		uid INTEGER NOT NULL,
		trip_id INTEGER DEFAULT 0,
		start_gaode_data TEXT,
		dest_gaode_data TEXT,
		depart_time TEXT DEFAULT '',
		time_type INTEGER DEFAULT 1,
		time_flex INTEGER DEFAULT 0,
		passenger_count INTEGER DEFAULT 1,
		is_charter INTEGER DEFAULT 0,
		order_status INTEGER DEFAULT 1,
		cancel_reason TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err := sqlite.DB.DB().Exec(sql)
	if err != nil {
		return err
	}
	return sqlite.DB.AutoMigrate(m.TableName(), &TbPpzOrder{})
}

// 新增订单
func (m *tbPpzOrderModel) Create(order *TbPpzOrder) (int64, error) {
	if order.OrderStatus == 0 {
		order.OrderStatus = OrderStatusMatching
	}
	if order.PassengerCount == 0 {
		order.PassengerCount = 1
	}
	if order.TimeType == 0 {
		order.TimeType = TimeTypeDepart
	}
	return sqlite.DB.Insert(m.TableName(), order)
}

// 按id查询
func (m *tbPpzOrderModel) GetById(orderId int64) (*TbPpzOrder, error) {
	var order TbPpzOrder
	order.OrderId = orderId
	err := sqlite.DB.GetByPkId(m.TableName(), &order)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// 更新
func (m *tbPpzOrderModel) Update(order *TbPpzOrder) error {
	order.UpdatedAt = time.Now()
	return sqlite.DB.UpdateByPkId(m.TableName(), order)
}

// 取消订单
func (m *tbPpzOrderModel) Cancel(orderId int64, reason string) error {
	order, err := m.GetById(orderId)
	if err != nil {
		return err
	}
	order.OrderStatus = OrderStatusCancelled
	order.CancelReason = reason
	return m.Update(order)
}

// 接单
func (m *tbPpzOrderModel) Accept(orderId int64, tripId int64) error {
	order, err := m.GetById(orderId)
	if err != nil {
		return err
	}
	order.OrderStatus = OrderStatusAccepted
	order.TripId = tripId
	return m.Update(order)
}

// 确认
func (m *tbPpzOrderModel) Confirm(orderId int64) error {
	order, err := m.GetById(orderId)
	if err != nil {
		return err
	}
	order.OrderStatus = OrderStatusConfirmed
	return m.Update(order)
}

// 匹配失败
func (m *tbPpzOrderModel) MatchFailed(orderId int64) error {
	order, err := m.GetById(orderId)
	if err != nil {
		return err
	}
	order.OrderStatus = OrderStatusMatchFailed
	return m.Update(order)
}

// 按用户查
func (m *tbPpzOrderModel) ListByUid(uid int64) ([]*TbPpzOrder, error) {
	var orders []*TbPpzOrder
	err := sqlite.DB.Find(m.TableName(), &orders, "uid = ? AND order_status > 0 ORDER BY created_at DESC", uid)
	return orders, err
}

// 按行程查
func (m *tbPpzOrderModel) ListByTripId(tripId int64) ([]*TbPpzOrder, error) {
	var orders []*TbPpzOrder
	err := sqlite.DB.Find(m.TableName(), &orders, "trip_id = ? AND order_status > 0 ORDER BY created_at ASC", tripId)
	return orders, err
}

// 按行程+用户查
func (m *tbPpzOrderModel) ListByTripIdAndUid(tripId, uid int64) ([]*TbPpzOrder, error) {
	var orders []*TbPpzOrder
	err := sqlite.DB.Find(m.TableName(), &orders, "trip_id = ? AND uid = ? AND order_status > 0 ORDER BY created_at ASC", tripId, uid)
	return orders, err
}

// 匹配中的订单列表
func (m *tbPpzOrderModel) ListMatching(limit, offset int) ([]*TbPpzOrder, error) {
	var orders []*TbPpzOrder
	err := sqlite.DB.Find(m.TableName(), &orders, "order_status = ? ORDER BY depart_date ASC, depart_time ASC LIMIT ? OFFSET ?", OrderStatusMatching, limit, offset)
	return orders, err
}

// 匹配中的数量
func (m *tbPpzOrderModel) CountMatching() (int64, error) {
	return sqlite.DB.Count(m.TableName(), "order_status = ?", OrderStatusMatching)
}

// 用户进行中的订单
func (m *tbPpzOrderModel) GetActiveByUid(uid int64) (*TbPpzOrder, error) {
	var order TbPpzOrder
	err := sqlite.DB.FindOne(m.TableName(), &order, "uid = ? AND order_status IN (?, ?, ?) ORDER BY created_at DESC LIMIT 1", uid, OrderStatusMatching, OrderStatusAccepted, OrderStatusConfirmed)
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// 用户最近的订单
func (m *tbPpzOrderModel) ListRecentByUid(uid int64, limit int) ([]*TbPpzOrder, error) {
	var orders []*TbPpzOrder
	err := sqlite.DB.Find(m.TableName(), &orders, "uid = ? AND order_status > 0 ORDER BY created_at DESC LIMIT ?", uid, limit)
	return orders, err
}

// 用户匹配中的订单列表
func (m *tbPpzOrderModel) ListMatchingByUid(uid int64) ([]*TbPpzOrder, error) {
	var orders []*TbPpzOrder
	err := sqlite.DB.Find(m.TableName(), &orders, "uid = ? AND order_status = ? ORDER BY created_at DESC", uid, OrderStatusMatching)
	return orders, err
}

// 用户已接单的行程ID列表（去重）
func (m *tbPpzOrderModel) ListActiveTripIdsByUid(uid int64) ([]int64, error) {
	rows, err := sqlite.DB.DB().Query("SELECT DISTINCT trip_id FROM tb_ppz_order WHERE uid = ? AND trip_id > 0 AND order_status IN (?, ?)", uid, OrderStatusAccepted, OrderStatusConfirmed)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
