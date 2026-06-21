package ppz

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/armylong/go-library/service/sqlite"
)

// 地址状态
const (
	AddressStatusDeleted = 0 // 已删除
	AddressStatusNormal  = 1 // 正常
)

// 常用地址
type TbPpzMapAddress struct {
	AddressId int64           `json:"address_id" db:"pk"` // 地址ID
	Uid       int64           `json:"uid"`                // 用户ID
	Remark    string          `json:"remark"`             // 地址备注
	GaodeData json.RawMessage `json:"gaode_data"`         // 高德地址数据
	Sort      int64           `json:"sort"`               // 排序值，越大越靠前
	Status    int             `json:"status"`             // 状态: 1-正常 0-删除
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

func (a TbPpzMapAddress) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *TbPpzMapAddress) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal TbPpzMapAddress value: %v", value)
	}
	return json.Unmarshal(bytes, a)
}

type tbPpzMapAddressModel struct{}

var TbPpzMapAddressModel = &tbPpzMapAddressModel{}

func init() {
	_ = TbPpzMapAddressModel.CreateTable()
}

func (m *tbPpzMapAddressModel) TableName() string {
	return "tb_ppz_map_address"
}

// 建表
func (m *tbPpzMapAddressModel) CreateTable() error {
	sql := `
	CREATE TABLE IF NOT EXISTS tb_ppz_map_address (
		address_id INTEGER PRIMARY KEY AUTOINCREMENT,
		uid INTEGER NOT NULL,
		remark TEXT DEFAULT '',
		gaode_data TEXT,
		sort INTEGER DEFAULT 0,
		status INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err := sqlite.DB.DB().Exec(sql)
	if err != nil {
		return err
	}
	return sqlite.DB.AutoMigrate(m.TableName(), &TbPpzMapAddress{})
}

// 新增地址
func (m *tbPpzMapAddressModel) Create(address *TbPpzMapAddress) (int64, error) {
	if address.Status == 0 {
		address.Status = AddressStatusNormal
	}
	if address.Sort == 0 {
		maxSort, _ := m.GetMaxSort(address.Uid)
		address.Sort = maxSort + 1
	}
	return sqlite.DB.Insert(m.TableName(), address)
}

// 按id查询
func (m *tbPpzMapAddressModel) GetById(addressId int64) (*TbPpzMapAddress, error) {
	var address TbPpzMapAddress
	address.AddressId = addressId
	err := sqlite.DB.GetByPkId(m.TableName(), &address)
	if err != nil {
		return nil, err
	}
	return &address, nil
}

// 按uid+addressId查询
func (m *tbPpzMapAddressModel) GetByUidAndId(uid, addressId int64) (*TbPpzMapAddress, error) {
	var address TbPpzMapAddress
	err := sqlite.DB.FindOne(m.TableName(), &address, "uid = ? AND address_id = ? AND status = ?", uid, addressId, AddressStatusNormal)
	if err != nil {
		return nil, err
	}
	return &address, nil
}

// 按用户查地址列表
func (m *tbPpzMapAddressModel) ListByUid(uid int64) ([]*TbPpzMapAddress, error) {
	var addresses []*TbPpzMapAddress
	err := sqlite.DB.Find(m.TableName(), &addresses, "uid = ? AND status = ? ORDER BY sort DESC, created_at DESC", uid, AddressStatusNormal)
	return addresses, err
}

// 获取用户最大排序值
func (m *tbPpzMapAddressModel) GetMaxSort(uid int64) (int64, error) {
	var maxSort int64
	err := sqlite.DB.DB().QueryRow("SELECT COALESCE(MAX(sort), 0) FROM "+m.TableName()+" WHERE uid = ? AND status = ?", uid, AddressStatusNormal).Scan(&maxSort)
	return maxSort, err
}

// 更新
func (m *tbPpzMapAddressModel) Update(address *TbPpzMapAddress) error {
	address.UpdatedAt = time.Now()
	return sqlite.DB.UpdateByPkId(m.TableName(), address)
}

// 删除（软删除）
func (m *tbPpzMapAddressModel) Delete(uid, addressId int64) error {
	address, err := m.GetByUidAndId(uid, addressId)
	if err != nil {
		return err
	}
	address.Status = AddressStatusDeleted
	return m.Update(address)
}

// 更新排序
func (m *tbPpzMapAddressModel) UpdateSort(uid, addressId int64, newSort int64) error {
	address, err := m.GetByUidAndId(uid, addressId)
	if err != nil {
		return err
	}
	address.Sort = newSort
	return m.Update(address)
}
