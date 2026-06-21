package user

import (
	"time"

	"github.com/armylong/go-library/service/sqlite"
)

// 管理员权限等级
const (
	UserPermissionNormal     = 0 // 普通用户
	UserPermissionAdmin      = 1 // 管理员
	UserPermissionSuperAdmin = 2 // 超级管理员
)

// 管理员用户
type TbAdminUser struct {
	ID         int64     `json:"id" db:"pk"` // 主键ID
	Uid        int64     `json:"uid"`        // 用户ID
	Permission int       `json:"permission"` // 权限等级
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type tbAdminUserModel struct{}

var TbAdminUserModel = &tbAdminUserModel{}

func init() {
	_ = TbAdminUserModel.CreateTable()
}

func (m *tbAdminUserModel) TableName() string {
	return "tb_admin_user"
}

// 建表
func (m *tbAdminUserModel) CreateTable() error {
	sql := `
	CREATE TABLE IF NOT EXISTS tb_admin_user (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uid INTEGER NOT NULL UNIQUE,
		permission INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err := sqlite.DB.DB().Exec(sql)
	if err != nil {
		return err
	}
	return sqlite.DB.AutoMigrate(m.TableName(), &TbAdminUser{})
}

// 新增
func (m *tbAdminUserModel) Create(admin *TbAdminUser) (int64, error) {
	return sqlite.DB.Insert(m.TableName(), admin)
}

// 按用户查
func (m *tbAdminUserModel) GetByUid(uid int64) (*TbAdminUser, error) {
	var admin TbAdminUser
	err := sqlite.DB.FindOne(m.TableName(), &admin, "uid = ?", uid)
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

// 获取用户权限等级
func (m *tbAdminUserModel) GetUserPermission(uid int64) int {
	admin, err := m.GetByUid(uid)
	if err != nil || admin == nil {
		return UserPermissionNormal
	}
	return admin.Permission
}

// 设置权限等级（不存在则创建）
func (m *tbAdminUserModel) SetPermission(uid int64, permission int) error {
	existing, _ := m.GetByUid(uid)
	if existing != nil {
		existing.Permission = permission
		return sqlite.DB.UpdateByPkId(m.TableName(), existing)
	}
	admin := &TbAdminUser{
		Uid:        uid,
		Permission: permission,
	}
	_, err := m.Create(admin)
	return err
}

// 设置/取消管理员
func (m *tbAdminUserModel) SetAdmin(uid int64, isAdmin bool) error {
	permission := UserPermissionNormal
	if isAdmin {
		permission = UserPermissionAdmin
	}
	return m.SetPermission(uid, permission)
}

// 管理员数量
func (m *tbAdminUserModel) CountAdmins() (int64, error) {
	return sqlite.DB.Count(m.TableName(), "permission >= ?", UserPermissionAdmin)
}

// 超级管理员数量
func (m *tbAdminUserModel) CountSuperAdmins() (int64, error) {
	return sqlite.DB.Count(m.TableName(), "permission = ?", UserPermissionSuperAdmin)
}

// 超级管理员列表
func (m *tbAdminUserModel) ListSuperAdmins() ([]*TbAdminUser, error) {
	var admins []*TbAdminUser
	err := sqlite.DB.Find(m.TableName(), &admins, "permission = ?", UserPermissionSuperAdmin)
	if err != nil {
		return nil, err
	}
	return admins, nil
}

// 所有管理员UID列表
func (m *tbAdminUserModel) ListAdminUids() ([]int64, error) {
	var admins []*TbAdminUser
	err := sqlite.DB.Find(m.TableName(), &admins, "permission >= ?", UserPermissionAdmin)
	if err != nil {
		return nil, err
	}
	uids := make([]int64, len(admins))
	for i, admin := range admins {
		uids[i] = admin.Uid
	}
	return uids, nil
}

// 所有管理员列表
func (m *tbAdminUserModel) ListAll() ([]*TbAdminUser, error) {
	var admins []*TbAdminUser
	err := sqlite.DB.Find(m.TableName(), &admins, "permission >= ?", UserPermissionAdmin)
	if err != nil {
		return nil, err
	}
	return admins, nil
}
