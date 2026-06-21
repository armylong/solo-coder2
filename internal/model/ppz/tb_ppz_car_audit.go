package ppz

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/armylong/go-library/service/sqlite"
)

// 审核状态
const (
	AuditStatusDeleted  = 0 // 已删除
	AuditStatusPending  = 1 // 待审核
	AuditStatusApproved = 2 // 已通过
	AuditStatusRejected = 3 // 已驳回
)

// 审核提交的车辆数据
type CarAuditData struct {
	CarModel           string `json:"car_model"`            // 车型
	CarLicensePhoto    string `json:"car_license_photo"`    // 行驶证照片
	DriverLicensePhoto string `json:"driver_license_photo"` // 驾驶证照片
	LicensePlate       string `json:"license_plate"`        // 车牌号
	CarColor           string `json:"car_color"`            // 车辆颜色
	Seats              int    `json:"seats"`                // 乘客座位数
	CarPhoto           string `json:"car_photo"`            // 车辆照片
	Description        string `json:"description"`          // 车辆简介
}

func (d CarAuditData) Value() (driver.Value, error) {
	return json.Marshal(d)
}

func (d *CarAuditData) Scan(value interface{}) error {
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
		return fmt.Errorf("cannot scan %T into CarAuditData", value)
	}
	return json.Unmarshal(bytes, d)
}

// 车辆审核记录
type TbPpzCarAudit struct {
	AuditId     int64         `json:"audit_id" db:"pk"` // 审核ID
	Uid         int64         `json:"uid"`              // 用户ID
	CarId       int64         `json:"car_id"`           // 车辆ID
	AuditData   *CarAuditData `json:"audit_data"`       // 审核的车辆数据
	AuditStatus int           `json:"audit_status"`     // 审核状态
	AuditReason string        `json:"audit_reason"`     // 审核原因
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type tbPpzCarAuditModel struct{}

var TbPpzCarAuditModel = &tbPpzCarAuditModel{}

func init() {
	_ = TbPpzCarAuditModel.CreateTable()
}

func (m *tbPpzCarAuditModel) TableName() string {
	return "tb_ppz_car_audit"
}

// 建表
func (m *tbPpzCarAuditModel) CreateTable() error {
	sql := `
	CREATE TABLE IF NOT EXISTS tb_ppz_car_audit (
		audit_id INTEGER PRIMARY KEY AUTOINCREMENT,
		uid INTEGER NOT NULL,
		car_id INTEGER DEFAULT 0,
		audit_data TEXT NOT NULL,
		audit_status INTEGER DEFAULT 1,
		audit_reason TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err := sqlite.DB.DB().Exec(sql)
	if err != nil {
		return err
	}
	return sqlite.DB.AutoMigrate(m.TableName(), &TbPpzCarAudit{})
}

// 新增审核记录
func (m *tbPpzCarAuditModel) Create(audit *TbPpzCarAudit) (int64, error) {
	if audit.AuditStatus == 0 {
		audit.AuditStatus = AuditStatusPending
	}
	return sqlite.DB.Insert(m.TableName(), audit)
}

// 按id查询
func (m *tbPpzCarAuditModel) GetById(auditId int64) (*TbPpzCarAudit, error) {
	var audit TbPpzCarAudit
	audit.AuditId = auditId
	err := sqlite.DB.GetByPkId(m.TableName(), &audit)
	if err != nil {
		return nil, err
	}
	return &audit, nil
}

// 按uid+auditId查询
func (m *tbPpzCarAuditModel) GetByUidAndId(uid, auditId int64) (*TbPpzCarAudit, error) {
	var audit TbPpzCarAudit
	err := sqlite.DB.FindOne(m.TableName(), &audit, "uid = ? AND audit_id = ?", uid, auditId)
	if err != nil {
		return nil, err
	}
	return &audit, nil
}

// 按uid+carId查最新一条
func (m *tbPpzCarAuditModel) GetByUidAndCarId(uid, carId int64) (*TbPpzCarAudit, error) {
	var audit TbPpzCarAudit
	err := sqlite.DB.FindOne(m.TableName(), &audit, "uid = ? AND car_id = ? AND audit_status > 0 ORDER BY created_at DESC LIMIT 1", uid, carId)
	if err != nil {
		return nil, err
	}
	return &audit, nil
}

// 按用户查审核列表
func (m *tbPpzCarAuditModel) ListByUid(uid int64, auditStatus int) ([]*TbPpzCarAudit, error) {
	var audits []*TbPpzCarAudit
	var err error

	if auditStatus == 0 {
		err = sqlite.DB.Find(m.TableName(), &audits, "uid = ? AND audit_status > 0 ORDER BY created_at DESC", uid)
	} else {
		err = sqlite.DB.Find(m.TableName(), &audits, "uid = ? AND audit_status = ? ORDER BY created_at DESC", uid, auditStatus)
	}

	return audits, err
}

// 更新
func (m *tbPpzCarAuditModel) Update(audit *TbPpzCarAudit) error {
	return sqlite.DB.UpdateByPkId(m.TableName(), audit)
}

// 删除（软删除）
func (m *tbPpzCarAuditModel) Delete(auditId int64) error {
	audit, err := m.GetById(auditId)
	if err != nil {
		return err
	}
	audit.AuditStatus = AuditStatusDeleted
	return m.Update(audit)
}

// 是否有审核通过的记录
func (m *tbPpzCarAuditModel) HasApprovedAudit(uid int64) (bool, error) {
	var audits []*TbPpzCarAudit
	err := sqlite.DB.Find(m.TableName(), &audits, "uid = ? AND audit_status = ? LIMIT 1", uid, AuditStatusApproved)
	if err != nil {
		return false, err
	}
	return len(audits) > 0, nil
}

// 按状态分页
func (m *tbPpzCarAuditModel) ListByStatus(auditStatus int, limit, offset int) ([]*TbPpzCarAudit, error) {
	var audits []*TbPpzCarAudit
	var err error

	if auditStatus == 0 {
		err = sqlite.DB.Find(m.TableName(), &audits, "audit_status > 0 ORDER BY updated_at ASC LIMIT ? OFFSET ?", limit, offset)
	} else {
		err = sqlite.DB.Find(m.TableName(), &audits, "audit_status = ? ORDER BY updated_at ASC LIMIT ? OFFSET ?", auditStatus, limit, offset)
	}

	return audits, err
}

// 按状态统计
func (m *tbPpzCarAuditModel) CountByStatus(auditStatus int) (int64, error) {
	if auditStatus == 0 {
		return sqlite.DB.Count(m.TableName(), "audit_status > 0")
	}
	return sqlite.DB.Count(m.TableName(), "audit_status = ?", auditStatus)
}

// 按状态获取去重uid列表
func (m *tbPpzCarAuditModel) GetDistinctUidsByStatus(auditStatus int) ([]int64, error) {
	var audits []*TbPpzCarAudit
	var err error

	if auditStatus == 0 {
		err = sqlite.DB.Find(m.TableName(), &audits, "audit_status > 0")
	} else {
		err = sqlite.DB.Find(m.TableName(), &audits, "audit_status = ?", auditStatus)
	}

	if err != nil {
		return nil, err
	}

	uidMap := make(map[int64]bool)
	for _, audit := range audits {
		uidMap[audit.Uid] = true
	}

	uids := make([]int64, 0, len(uidMap))
	for uid := range uidMap {
		uids = append(uids, uid)
	}

	return uids, nil
}

// 按用户+状态查审核列表
func (m *tbPpzCarAuditModel) ListByUidAndStatus(uid int64, auditStatus int) ([]*TbPpzCarAudit, error) {
	var audits []*TbPpzCarAudit
	var err error

	if auditStatus == 0 {
		err = sqlite.DB.Find(m.TableName(), &audits, "uid = ? AND audit_status > 0 ORDER BY updated_at ASC", uid)
	} else {
		err = sqlite.DB.Find(m.TableName(), &audits, "uid = ? AND audit_status = ? ORDER BY updated_at ASC", uid, auditStatus)
	}

	return audits, err
}
