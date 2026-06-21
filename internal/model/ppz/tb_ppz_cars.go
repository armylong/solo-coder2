package ppz

import (
	"time"

	"github.com/armylong/go-library/service/sqlite"
)

// 车辆状态
const (
	CarStatusDeleted = 0 // 已删除
	CarStatusNormal  = 1 // 正常
)

// 车辆
type TbPpzCars struct {
	CarId              int64     `json:"car_id" db:"pk"`          // 车辆ID
	Uid                int64     `json:"uid"`                     // 所属用户ID
	CarModel           string    `json:"car_model"`               // 车型
	CarLicensePhoto    string    `json:"car_license_photo"`       // 行驶证照片
	DriverLicensePhoto string    `json:"driver_license_photo"`    // 驾驶证照片
	LicensePlate       string    `json:"license_plate"`           // 车牌号
	CarColor           string    `json:"car_color"`               // 车辆颜色
	Seats              int       `json:"seats"`                   // 乘客座位数
	CarPhoto           string    `json:"car_photo"`               // 车辆照片
	Description        string    `json:"description"`             // 车辆简介
	Status             int       `json:"status"`                  // 状态: 1-正常 0-删除
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type tbPpzCarsModel struct{}

var TbPpzCarsModel = &tbPpzCarsModel{}

func init() {
	_ = TbPpzCarsModel.CreateTable()
}

func (m *tbPpzCarsModel) TableName() string {
	return "tb_ppz_cars"
}

// 建表
func (m *tbPpzCarsModel) CreateTable() error {
	sql := `
	CREATE TABLE IF NOT EXISTS tb_ppz_cars (
		car_id INTEGER PRIMARY KEY AUTOINCREMENT,
		uid INTEGER NOT NULL,
		car_model TEXT NOT NULL,
		car_license_photo TEXT NOT NULL,
		driver_license_photo TEXT NOT NULL,
		license_plate TEXT NOT NULL,
		car_color TEXT NOT NULL,
		seats INTEGER NOT NULL DEFAULT 5,
		car_photo TEXT NOT NULL,
		description TEXT NOT NULL,
		status INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err := sqlite.DB.DB().Exec(sql)
	if err != nil {
		return err
	}
	return sqlite.DB.AutoMigrate(m.TableName(), &TbPpzCars{})
}

// 新增车辆
func (m *tbPpzCarsModel) Create(car *TbPpzCars) (int64, error) {
	if car.Status == 0 {
		car.Status = CarStatusNormal
	}
	return sqlite.DB.Insert(m.TableName(), car)
}

// 按id查询
func (m *tbPpzCarsModel) GetById(carId int64) (*TbPpzCars, error) {
	var car TbPpzCars
	car.CarId = carId
	err := sqlite.DB.GetByPkId(m.TableName(), &car)
	if err != nil {
		return nil, err
	}
	return &car, nil
}

// 按uid+carId查询
func (m *tbPpzCarsModel) GetByUidAndId(uid, carId int64) (*TbPpzCars, error) {
	var car TbPpzCars
	err := sqlite.DB.FindOne(m.TableName(), &car, "uid = ? AND car_id = ?", uid, carId)
	if err != nil {
		return nil, err
	}
	return &car, nil
}

// 按用户查车辆列表
func (m *tbPpzCarsModel) ListByUid(uid int64, status int) ([]*TbPpzCars, error) {
	var cars []*TbPpzCars
	var err error

	if status == 0 {
		err = sqlite.DB.Find(m.TableName(), &cars, "uid = ? AND status = 1 ORDER BY created_at DESC", uid)
	} else {
		err = sqlite.DB.Find(m.TableName(), &cars, "uid = ? AND status = ? ORDER BY created_at DESC", uid, status)
	}

	return cars, err
}

// 删除
func (m *tbPpzCarsModel) Delete(carId int64) error {
	car := &TbPpzCars{CarId: carId}
	return sqlite.DB.DeleteByPkId(m.TableName(), car)
}

// 更新
func (m *tbPpzCarsModel) Update(car *TbPpzCars) error {
	return sqlite.DB.UpdateByPkId(m.TableName(), car)
}

// 车辆总数
func (m *tbPpzCarsModel) CountCars() (int64, error) {
	return sqlite.DB.CountAll(m.TableName())
}

// 按用户统计车辆数
func (m *tbPpzCarsModel) CountByUid(uid int64) (int64, error) {
	var cars []*TbPpzCars
	err := sqlite.DB.Find(m.TableName(), &cars, "uid = ? AND status = 1", uid)
	if err != nil {
		return 0, err
	}
	return int64(len(cars)), nil
}

// 按用户查所有车辆
func (m *tbPpzCarsModel) ListAllByUid(uid int64) ([]*TbPpzCars, error) {
	var cars []*TbPpzCars
	err := sqlite.DB.Find(m.TableName(), &cars, "uid = ? AND status = 1 ORDER BY created_at DESC", uid)
	return cars, err
}

// 所有去重uid
func (m *tbPpzCarsModel) GetAllDistinctUids() ([]int64, error) {
	var cars []*TbPpzCars
	err := sqlite.DB.Find(m.TableName(), &cars, "1=1")
	if err != nil {
		return nil, err
	}

	uidMap := make(map[int64]bool)
	for _, car := range cars {
		uidMap[car.Uid] = true
	}

	uids := make([]int64, 0, len(uidMap))
	for uid := range uidMap {
		uids = append(uids, uid)
	}

	return uids, nil
}

// 按司机状态筛选uid
func (m *tbPpzCarsModel) GetDistinctUidsByStatus(status int) ([]int64, error) {
	allUids, err := m.GetAllDistinctUids()
	if err != nil {
		return nil, err
	}

	if status == 0 {
		return allUids, nil
	}

	var filteredUids []int64
	for _, uid := range allUids {
		isDriverBanned, _ := TbPpzUserModel.IsDriverBanned(uid)
		if status == 2 && isDriverBanned {
			filteredUids = append(filteredUids, uid)
		} else if status == 1 && !isDriverBanned {
			filteredUids = append(filteredUids, uid)
		}
	}

	return filteredUids, nil
}

// 去重uid数
func (m *tbPpzCarsModel) CountDistinctUids() (int64, error) {
	uids, err := m.GetAllDistinctUids()
	if err != nil {
		return 0, err
	}
	return int64(len(uids)), nil
}
