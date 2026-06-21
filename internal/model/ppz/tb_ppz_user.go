package ppz

import (
	"time"

	"github.com/armylong/go-library/service/sqlite"
)

// 用户状态
const (
	UserStatusNormal = 1 // 正常
	UserStatusBanned = 2 // 已封禁
)

// 司机状态
const (
	DriverStatusNormal = 1 // 正常
	DriverStatusBanned = 2 // 已封禁
)

// 用户扩展信息
type TbPpzUser struct {
	Id           int64     `json:"id" db:"pk"`            // 主键
	Uid          int64     `json:"uid"`                   // 用户ID
	CarCount     int       `json:"car_count"`             // 车辆数，>0为司机
	Status       int       `json:"status"`                // 用户状态: 1-正常 2-已封禁
	DriverStatus int       `json:"driver_status"`         // 司机状态: 1-正常 2-已封禁
	BanReason    string    `json:"ban_reason,omitempty"`  // 封禁原因
	BannedAt     time.Time `json:"banned_at,omitempty"`   // 封禁时间
	UnbannedAt   time.Time `json:"unbanned_at,omitempty"` // 解封时间
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type tbPpzUserModel struct{}

var TbPpzUserModel = &tbPpzUserModel{}

func init() {
	_ = TbPpzUserModel.CreateTable()
}

func (m *tbPpzUserModel) TableName() string {
	return "tb_ppz_user"
}

// 建表
func (m *tbPpzUserModel) CreateTable() error {
	sql := `
	CREATE TABLE IF NOT EXISTS tb_ppz_user (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uid INTEGER NOT NULL UNIQUE,
		car_count INTEGER DEFAULT 0,
		status INTEGER DEFAULT 1,
		driver_status INTEGER DEFAULT 1,
		ban_reason TEXT DEFAULT '',
		banned_at DATETIME,
		unbanned_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err := sqlite.DB.DB().Exec(sql)
	if err != nil {
		return err
	}
	return sqlite.DB.AutoMigrate(m.TableName(), &TbPpzUser{})
}

// 新增用户
func (m *tbPpzUserModel) Create(user *TbPpzUser) (int64, error) {
	if user.Status == 0 {
		user.Status = UserStatusNormal
	}
	if user.DriverStatus == 0 {
		user.DriverStatus = DriverStatusNormal
	}
	if user.CarCount < 0 {
		user.CarCount = 0
	}
	return sqlite.DB.Insert(m.TableName(), user)
}

// 按uid查询
func (m *tbPpzUserModel) GetByUid(uid int64) (*TbPpzUser, error) {
	var user TbPpzUser
	err := sqlite.DB.FindOne(m.TableName(), &user, "uid = ?", uid)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// 获取或创建
func (m *tbPpzUserModel) GetOrCreateByUid(uid int64) (*TbPpzUser, error) {
	user, err := m.GetByUid(uid)
	if err == nil && user != nil {
		return user, nil
	}

	user = &TbPpzUser{
		Uid:          uid,
		CarCount:     0,
		Status:       UserStatusNormal,
		DriverStatus: DriverStatusNormal,
	}
	_, err = m.Create(user)
	if err != nil {
		return nil, err
	}
	return m.GetByUid(uid)
}

// 更新
func (m *tbPpzUserModel) Update(user *TbPpzUser) error {
	return sqlite.DB.UpdateByPkId(m.TableName(), user)
}

// 用户是否被封禁
func (m *tbPpzUserModel) IsUserBanned(uid int64) (bool, error) {
	user, err := m.GetByUid(uid)
	if err != nil {
		return false, nil
	}
	return user.Status == UserStatusBanned, nil
}

// 司机是否被封禁
func (m *tbPpzUserModel) IsDriverBanned(uid int64) (bool, error) {
	user, err := m.GetByUid(uid)
	if err != nil {
		return false, nil
	}
	return user.DriverStatus == DriverStatusBanned, nil
}

// 是否是司机
func (m *tbPpzUserModel) IsDriver(uid int64) (bool, error) {
	user, err := m.GetByUid(uid)
	if err != nil {
		return false, nil
	}
	return user.CarCount > 0, nil
}

// 增加车辆数
func (m *tbPpzUserModel) IncrCarCount(uid int64, delta int) error {
	user, err := m.GetOrCreateByUid(uid)
	if err != nil {
		return err
	}

	user.CarCount += delta
	if user.CarCount < 0 {
		user.CarCount = 0
	}
	return m.Update(user)
}

// 减少车辆数
func (m *tbPpzUserModel) DecrCarCount(uid int64, delta int) error {
	return m.IncrCarCount(uid, -delta)
}

// 封禁用户
func (m *tbPpzUserModel) BanUser(uid int64, reason string) error {
	user, err := m.GetOrCreateByUid(uid)
	if err != nil {
		return err
	}

	user.Status = UserStatusBanned
	user.BanReason = reason
	user.BannedAt = time.Now()
	return m.Update(user)
}

// 解封用户
func (m *tbPpzUserModel) UnbanUser(uid int64) error {
	user, err := m.GetByUid(uid)
	if err != nil {
		return err
	}

	user.Status = UserStatusNormal
	user.UnbannedAt = time.Now()
	return m.Update(user)
}

// 封禁司机
func (m *tbPpzUserModel) BanDriver(uid int64, reason string) error {
	user, err := m.GetOrCreateByUid(uid)
	if err != nil {
		return err
	}

	user.DriverStatus = DriverStatusBanned
	user.BanReason = reason
	user.BannedAt = time.Now()
	return m.Update(user)
}

// 解封司机
func (m *tbPpzUserModel) UnbanDriver(uid int64) error {
	user, err := m.GetByUid(uid)
	if err != nil {
		return err
	}

	user.DriverStatus = DriverStatusNormal
	user.UnbannedAt = time.Now()
	return m.Update(user)
}

// 用户总数
func (m *tbPpzUserModel) CountUsers() (int64, error) {
	return sqlite.DB.CountAll(m.TableName())
}

// 封禁用户数
func (m *tbPpzUserModel) CountBannedUsers() (int64, error) {
	return sqlite.DB.Count(m.TableName(), "status = 2")
}

// 司机数
func (m *tbPpzUserModel) CountDrivers() (int64, error) {
	return sqlite.DB.Count(m.TableName(), "car_count > 0")
}

// 分页列表
func (m *tbPpzUserModel) List(limit, offset int) ([]*TbPpzUser, error) {
	var users []*TbPpzUser
	err := sqlite.DB.Find(m.TableName(), &users, "1=1 ORDER BY updated_at DESC LIMIT ? OFFSET ?", limit, offset)
	return users, err
}

// 按状态分页
func (m *tbPpzUserModel) ListByStatus(status int, limit, offset int) ([]*TbPpzUser, error) {
	var users []*TbPpzUser
	var err error

	if status == 0 {
		err = sqlite.DB.Find(m.TableName(), &users, "1=1 ORDER BY updated_at DESC LIMIT ? OFFSET ?", limit, offset)
	} else {
		err = sqlite.DB.Find(m.TableName(), &users, "status = ? ORDER BY updated_at DESC LIMIT ? OFFSET ?", status, limit, offset)
	}
	return users, err
}
