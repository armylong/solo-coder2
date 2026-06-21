package long_doc

import (
	"database/sql"
	"time"

	"github.com/armylong/go-library/service/sqlite"
)

// 长文档
type TbLongDoc struct {
	DocId       int64     `json:"doc_id" db:"pk"`  // 文档ID
	Uid         int64     `json:"uid"`             // 用户ID
	ParentDocId int64     `json:"parent_doc_id"`   // 父文档ID
	DocName     string    `json:"doc_name"`        // 文档名称
	DocValue    string    `json:"doc_value"`       // 文档内容
	SortOrder   int       `json:"sort_order"`      // 排序值
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type tbLongDocModel struct{}

var TbLongDocModel = &tbLongDocModel{}

func init() {
	_ = TbLongDocModel.CreateTable()
}

func (m *tbLongDocModel) TableName() string {
	return "tb_long_doc"
}

// 建表
func (m *tbLongDocModel) CreateTable() error {
	sql := `
	CREATE TABLE IF NOT EXISTS tb_long_doc (
		doc_id INTEGER PRIMARY KEY AUTOINCREMENT,
		uid INTEGER NOT NULL,
		parent_doc_id INTEGER DEFAULT 0,
		doc_name TEXT NOT NULL DEFAULT '',
		doc_value TEXT,
		sort_order INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err := sqlite.DB.DB().Exec(sql)
	if err != nil {
		return err
	}
	return sqlite.DB.AutoMigrate(m.TableName(), &TbLongDoc{})
}

// 创建文档
func (m *tbLongDocModel) Create(data *TbLongDoc) (int64, error) {
	if data.SortOrder == 0 {
		maxOrder, _ := m.GetMaxSortOrder(data.Uid, data.ParentDocId)
		data.SortOrder = maxOrder + 1
	}
	return sqlite.DB.Insert(m.TableName(), data)
}

// 根据ID获取
func (m *tbLongDocModel) GetById(docId int64) (*TbLongDoc, error) {
	var row TbLongDoc
	row.DocId = docId
	err := sqlite.DB.GetByPkId(m.TableName(), &row)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// 根据ID和用户ID获取
func (m *tbLongDocModel) GetByIdAndUid(docId, uid int64) (*TbLongDoc, error) {
	var row TbLongDoc
	err := sqlite.DB.FindOne(m.TableName(), &row, "doc_id = ? AND uid = ?", docId, uid)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// 获取用户所有文档
func (m *tbLongDocModel) ListByUid(uid int64) ([]*TbLongDoc, error) {
	var list []*TbLongDoc
	err := sqlite.DB.Find(m.TableName(), &list, "uid = ? ORDER BY parent_doc_id ASC, sort_order ASC", uid)
	return list, err
}

// 获取父文档下的子文档
func (m *tbLongDocModel) ListByParentId(uid, parentDocId int64) ([]*TbLongDoc, error) {
	var list []*TbLongDoc
	err := sqlite.DB.Find(m.TableName(), &list, "uid = ? AND parent_doc_id = ? ORDER BY sort_order ASC", uid, parentDocId)
	return list, err
}

// 更新文档
func (m *tbLongDocModel) Update(data *TbLongDoc) error {
	data.UpdatedAt = time.Now()
	return sqlite.DB.UpdateByPkId(m.TableName(), data)
}

// 更新文档内容
func (m *tbLongDocModel) UpdateDocValue(docId, uid int64, docValue string) error {
	sql := "UPDATE tb_long_doc SET doc_value = ?, updated_at = CURRENT_TIMESTAMP WHERE doc_id = ? AND uid = ?"
	_, err := sqlite.DB.DB().Exec(sql, docValue, docId, uid)
	return err
}

// 更新父文档和排序
func (m *tbLongDocModel) UpdateParentAndSort(docId, uid, parentDocId int64, sortOrder int) error {
	sql := "UPDATE tb_long_doc SET parent_doc_id = ?, sort_order = ?, updated_at = CURRENT_TIMESTAMP WHERE doc_id = ? AND uid = ?"
	_, err := sqlite.DB.DB().Exec(sql, parentDocId, sortOrder, docId, uid)
	return err
}

// 获取同级最大排序值
func (m *tbLongDocModel) GetMaxSortOrder(uid, parentDocId int64) (int, error) {
	var maxOrder int
	sql := "SELECT IFNULL(MAX(sort_order), 0) FROM tb_long_doc WHERE uid = ? AND parent_doc_id = ?"
	err := sqlite.DB.DB().QueryRow(sql, uid, parentDocId).Scan(&maxOrder)
	if err != nil {
		return 0, err
	}
	return maxOrder, nil
}

// 递归删除文档及子文档
func (m *tbLongDocModel) Delete(docId, uid int64) error {
	childDocs, err := m.ListByParentId(uid, docId)
	if err != nil {
		return err
	}
	
	for _, child := range childDocs {
		if err := m.Delete(child.DocId, uid); err != nil {
			return err
		}
	}
	
	data := &TbLongDoc{DocId: docId}
	return sqlite.DB.DeleteByPkId(m.TableName(), data)
}

func (m *tbLongDocModel) GetDB() *sql.DB {
	return sqlite.DB.DB()
}
