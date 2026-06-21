package user

import (
	"time"

	"github.com/armylong/go-library/service/sqlite"
)

// 用户
type TbUser struct {
	Uid       int64     `json:"uid" db:"pk"` // 主键ID
	Account   string    `json:"account"`     // 账号
	Password  string    `json:"password"`    // 密码
	Name      string    `json:"name"`        // 用户名
	Email     string    `json:"email"`       // 邮箱
	Phone     string    `json:"phone"`       // 手机号
	Status    int       `json:"status"`      // 状态: 1-正常 0-禁用
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// 清空密码（返回前端前调用）
func (u *TbUser) ClearPassword() {
	u.Password = ""
}

type tbUserModel struct{}

var TbUserModel = &tbUserModel{}

func init() {
	_ = TbUserModel.CreateTable()
}

func (m *tbUserModel) TableName() string {
	return "tb_user"
}

// 建表
func (m *tbUserModel) CreateTable() error {
	sql := `
	CREATE TABLE IF NOT EXISTS tb_user (
		uid INTEGER PRIMARY KEY AUTOINCREMENT,
		account TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		name TEXT NOT NULL,
		email TEXT,
		phone TEXT,
		status INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err := sqlite.DB.DB().Exec(sql)
	if err != nil {
		return err
	}
	return sqlite.DB.AutoMigrate(m.TableName(), &TbUser{})
}

// 新增
func (m *tbUserModel) Create(user *TbUser) (int64, error) {
	return sqlite.DB.Insert(m.TableName(), user)
}

// 按用户ID查
func (m *tbUserModel) GetByUid(uid int64) (*TbUser, error) {
	var user TbUser
	user.Uid = uid
	err := sqlite.DB.GetByPkId(m.TableName(), &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// 按邮箱查
func (m *tbUserModel) GetByEmail(email string) (*TbUser, error) {
	var user TbUser
	err := sqlite.DB.FindOne(m.TableName(), &user, "email = ?", email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// 按账号查
func (m *tbUserModel) GetByAccount(account string) (*TbUser, error) {
	var user TbUser
	err := sqlite.DB.FindOne(m.TableName(), &user, "account = ?", account)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// 分页列表
func (m *tbUserModel) List(limit, offset int) ([]*TbUser, error) {
	var users []*TbUser
	err := sqlite.DB.Find(m.TableName(), &users, "1=1 ORDER BY uid DESC LIMIT ? OFFSET ?", limit, offset)
	return users, err
}

// 更新
func (m *tbUserModel) Update(user *TbUser) error {
	return sqlite.DB.UpdateByPkId(m.TableName(), user)
}

// 删除
func (m *tbUserModel) Delete(id int64) error {
	user := &TbUser{Uid: id}
	return sqlite.DB.DeleteByPkId(m.TableName(), user)
}

// 总数
func (m *tbUserModel) Count() (int64, error) {
	return sqlite.DB.CountAll(m.TableName())
}

// 搜索用户（按账号、昵称或手机号模糊搜索）
func (m *tbUserModel) Search(keyword string, limit int) ([]*TbUser, error) {
	var users []*TbUser
	pattern := "%" + keyword + "%"
	err := sqlite.DB.Find(m.TableName(), &users,
		"(account LIKE ? OR name LIKE ? OR phone LIKE ?) AND status = 1 ORDER BY uid DESC LIMIT ?",
		pattern, pattern, pattern, limit,
	)
	return users, err
}
