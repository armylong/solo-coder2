package ppz

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/armylong/go-library/service/sqlite"
)

// 审核日志
type TbPpzCarAuditLog struct {
	AuditLogId  int64         `json:"audit_log_id" db:"pk"` // 日志ID
	AuditId     int64         `json:"audit_id"`             // 审核记录ID
	Uid         int64         `json:"uid"`                  // 用户ID
	CarId       int64         `json:"car_id"`               // 车辆ID
	AuditData   *CarAuditData `json:"audit_data"`           // 审核的车辆数据
	AuditStatus int           `json:"audit_status"`         // 审核状态
	AuditReason string        `json:"audit_reason"`         // 审核原因
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type tbPpzCarAuditLogModel struct{}

var TbPpzCarAuditLogModel = &tbPpzCarAuditLogModel{}

func init() {
	_ = TbPpzCarAuditLogModel.CreateTable()
}

func (m *tbPpzCarAuditLogModel) TableName() string {
	return "tb_ppz_car_audit_log"
}

// 建表
func (m *tbPpzCarAuditLogModel) CreateTable() error {
	sql := `
	CREATE TABLE IF NOT EXISTS tb_ppz_car_audit_log (
		audit_log_id INTEGER PRIMARY KEY AUTOINCREMENT,
		audit_id INTEGER NOT NULL,
		uid INTEGER NOT NULL,
		car_id INTEGER DEFAULT 0,
		audit_data TEXT NOT NULL,
		audit_status INTEGER NOT NULL,
		audit_reason TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err := sqlite.DB.DB().Exec(sql)
	if err != nil {
		return err
	}
	return sqlite.DB.AutoMigrate(m.TableName(), &TbPpzCarAuditLog{})
}

// 新增日志
func (m *tbPpzCarAuditLogModel) Create(log *TbPpzCarAuditLog) (int64, error) {
	return sqlite.DB.Insert(m.TableName(), log)
}

// 按id查询
func (m *tbPpzCarAuditLogModel) GetById(auditLogId int64) (*TbPpzCarAuditLog, error) {
	var log TbPpzCarAuditLog
	log.AuditLogId = auditLogId
	err := sqlite.DB.GetByPkId(m.TableName(), &log)
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// 按审核id查日志列表
func (m *tbPpzCarAuditLogModel) ListByAuditId(auditId int64) ([]*TbPpzCarAuditLog, error) {
	var logs []*TbPpzCarAuditLog
	err := sqlite.DB.Find(m.TableName(), &logs, "audit_id = ? ORDER BY created_at DESC", auditId)
	return logs, err
}

// 按用户分页查
func (m *tbPpzCarAuditLogModel) ListByUid(uid int64, limit, offset int) ([]*TbPpzCarAuditLog, error) {
	var logs []*TbPpzCarAuditLog
	err := sqlite.DB.Find(m.TableName(), &logs, "uid = ? ORDER BY created_at DESC LIMIT ? OFFSET ?", uid, limit, offset)
	return logs, err
}

// 从审核记录创建日志
func (m *tbPpzCarAuditLogModel) CreateFromAudit(audit *TbPpzCarAudit) (int64, error) {
	log := &TbPpzCarAuditLog{
		AuditId:     audit.AuditId,
		Uid:         audit.Uid,
		CarId:       audit.CarId,
		AuditData:   audit.AuditData,
		AuditStatus: audit.AuditStatus,
		AuditReason: audit.AuditReason,
	}
	return m.Create(log)
}

// 审核数据转字符串
func auditDataToString(data *CarAuditData) string {
	if data == nil {
		return "{}"
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return "{}"
	}
	return string(bytes)
}

// 字符串转审核数据
func stringToAuditData(dataStr string) (*CarAuditData, error) {
	if dataStr == "" {
		return &CarAuditData{}, nil
	}
	var data CarAuditData
	err := json.Unmarshal([]byte(dataStr), &data)
	if err != nil {
		return nil, fmt.Errorf("解析audit_data失败: %w", err)
	}
	return &data, nil
}
