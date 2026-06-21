package api_catcher

import (
	"database/sql"
	"time"

	"github.com/armylong/go-library/service/sqlite"
)

// API抓包记录
type TbApiCatcher struct {
	ID        int64     `json:"id" db:"pk"` // 主键ID
	Data      string    `json:"data"`       // 抓包数据(JSON)
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type tbApiCatcherModel struct{}

var TbApiCatcherModel = &tbApiCatcherModel{}

func init() {
	_ = TbApiCatcherModel.CreateTable()
}

func (m *tbApiCatcherModel) TableName() string {
	return "tb_api_catcher"
}

// 建表
func (m *tbApiCatcherModel) CreateTable() error {
	sql := `
	CREATE TABLE IF NOT EXISTS tb_api_catcher (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		data TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err := sqlite.DB.DB().Exec(sql)
	if err != nil {
		return err
	}
	return sqlite.DB.AutoMigrate(m.TableName(), &TbApiCatcher{})
}

// 新增
func (m *tbApiCatcherModel) Create(data *TbApiCatcher) (int64, error) {
	return sqlite.DB.Insert(m.TableName(), data)
}

// 按ID查
func (m *tbApiCatcherModel) GetById(id int64) (*TbApiCatcher, error) {
	var row TbApiCatcher
	row.ID = id
	err := sqlite.DB.GetByPkId(m.TableName(), &row)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// 分页列表
func (m *tbApiCatcherModel) List(limit, offset int) ([]*TbApiCatcher, error) {
	var list []*TbApiCatcher
	err := sqlite.DB.Find(m.TableName(), &list, "1=1 ORDER BY id DESC LIMIT ? OFFSET ?", limit, offset)
	return list, err
}

// 总数
func (m *tbApiCatcherModel) Count() (int64, error) {
	return sqlite.DB.CountAll(m.TableName())
}

// 按ID删除
func (m *tbApiCatcherModel) Delete(id int64) error {
	data := &TbApiCatcher{ID: id}
	return sqlite.DB.DeleteByPkId(m.TableName(), data)
}

// 按日期下载，支持从指定ID之后续传
func (m *tbApiCatcherModel) Download(id int64, limit int, date string) ([]*TbApiCatcher, error) {
	var list []*TbApiCatcher

	if limit <= 0 {
		limit = 50
	}
	if limit > 50 {
		limit = 50
	}

	var err error
	if id > 0 {
		err = sqlite.DB.Find(m.TableName(), &list, "id > ? AND date(created_at) = ? ORDER BY id DESC LIMIT ?", id, date, limit)
	} else {
		err = sqlite.DB.Find(m.TableName(), &list, "date(created_at) = ? ORDER BY id DESC LIMIT ?", date, limit)
	}
	return list, err
}

// 获取原始DB连接
func (m *tbApiCatcherModel) GetDB() *sql.DB {
	return sqlite.DB.DB()
}
