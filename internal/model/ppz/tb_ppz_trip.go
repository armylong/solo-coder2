package ppz

import (
	"time"

	"github.com/armylong/go-library/service/sqlite"
)

// 行程状态
const (
	TripStatusCancelled  = 0 // 已取消
	TripStatusPreparing  = 1 // 准备中
	TripStatusInProgress = 2 // 进行中
	TripStatusCompleted  = 3 // 已完成
)

// 行程
type TbPpzTrip struct {
	TripId    int64     `json:"trip_id" db:"pk"` // 行程ID
	DriverId  int64     `json:"driver_id"`       // 司机用户ID
	Status    int       `json:"status"`          // 行程状态
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type tbPpzTripModel struct{}

var TbPpzTripModel = &tbPpzTripModel{}

func init() {
	_ = TbPpzTripModel.CreateTable()
}

func (m *tbPpzTripModel) TableName() string {
	return "tb_ppz_trip"
}

// 建表
func (m *tbPpzTripModel) CreateTable() error {
	sql := `
	CREATE TABLE IF NOT EXISTS tb_ppz_trip (
		trip_id INTEGER PRIMARY KEY AUTOINCREMENT,
		driver_id INTEGER NOT NULL,
		status INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err := sqlite.DB.DB().Exec(sql)
	if err != nil {
		return err
	}
	return sqlite.DB.AutoMigrate(m.TableName(), &TbPpzTrip{})
}

// 新增行程
func (m *tbPpzTripModel) Create(trip *TbPpzTrip) (int64, error) {
	if trip.Status == 0 {
		trip.Status = TripStatusPreparing
	}
	return sqlite.DB.Insert(m.TableName(), trip)
}

// 按id查询
func (m *tbPpzTripModel) GetById(tripId int64) (*TbPpzTrip, error) {
	var trip TbPpzTrip
	trip.TripId = tripId
	err := sqlite.DB.GetByPkId(m.TableName(), &trip)
	if err != nil {
		return nil, err
	}
	return &trip, nil
}

// 更新
func (m *tbPpzTripModel) Update(trip *TbPpzTrip) error {
	trip.UpdatedAt = time.Now()
	return sqlite.DB.UpdateByPkId(m.TableName(), trip)
}

// 取消行程
func (m *tbPpzTripModel) Cancel(tripId int64) error {
	trip, err := m.GetById(tripId)
	if err != nil {
		return err
	}
	trip.Status = TripStatusCancelled
	return m.Update(trip)
}

// 开始行程
func (m *tbPpzTripModel) StartTrip(tripId int64) error {
	trip, err := m.GetById(tripId)
	if err != nil {
		return err
	}
	trip.Status = TripStatusInProgress
	return m.Update(trip)
}

// 完成行程
func (m *tbPpzTripModel) CompleteTrip(tripId int64) error {
	trip, err := m.GetById(tripId)
	if err != nil {
		return err
	}
	trip.Status = TripStatusCompleted
	return m.Update(trip)
}

// 司机进行中的行程
func (m *tbPpzTripModel) GetActiveByDriverId(driverId int64) (*TbPpzTrip, error) {
	var trip TbPpzTrip
	err := sqlite.DB.FindOne(m.TableName(), &trip, "driver_id = ? AND status IN (?, ?) ORDER BY created_at DESC LIMIT 1", driverId, TripStatusPreparing, TripStatusInProgress)
	if err != nil {
		return nil, err
	}
	return &trip, nil
}

// 按司机分页查
func (m *tbPpzTripModel) ListByDriverId(driverId int64, limit, offset int) ([]*TbPpzTrip, error) {
	var trips []*TbPpzTrip
	err := sqlite.DB.Find(m.TableName(), &trips, "driver_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?", driverId, limit, offset)
	return trips, err
}
