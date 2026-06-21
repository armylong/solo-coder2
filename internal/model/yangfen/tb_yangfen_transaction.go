package yangfen

import (
	"time"

	"github.com/armylong/go-library/service/sqlite"
)

// 交易类型
const (
	TxTypeRecharge     = "recharge"     // 充值
	TxTypeConsume      = "consume"      // 消费
	TxTypeTransferOut  = "transfer_out" // 转出
	TxTypeTransferIn   = "transfer_in"  // 转入
	TxTypeRefund       = "refund"       // 退款
)

// 氧分交易记录
type TbYangfenTransaction struct {
	ID            int64     `json:"id" db:"pk"`     // 主键ID
	TransactionId string    `json:"transaction_id"` // 交易号
	Uid           string    `json:"uid"`            // 用户ID
	Type          string    `json:"type"`           // 交易类型
	Amount        int       `json:"amount"`         // 交易金额
	Balance       int       `json:"balance"`        // 交易后余额
	Description   string    `json:"description"`    // 交易描述
	CreatedAt     time.Time `json:"created_at"`
}

type tbYangfenTransactionModel struct{}

var TbYangfenTransactionModel = &tbYangfenTransactionModel{}

func init() {
	_ = TbYangfenTransactionModel.CreateTable()
}

func (m *tbYangfenTransactionModel) TableName() string {
	return "tb_yangfen_transaction"
}

// 建表
func (m *tbYangfenTransactionModel) CreateTable() error {
	sql := `
	CREATE TABLE IF NOT EXISTS tb_yangfen_transaction (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		transaction_id TEXT NOT NULL UNIQUE,
		uid TEXT NOT NULL,
		type TEXT NOT NULL,
		amount INTEGER DEFAULT 0,
		balance INTEGER DEFAULT 0,
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err := sqlite.DB.DB().Exec(sql)
	if err != nil {
		return err
	}
	return sqlite.DB.AutoMigrate(m.TableName(), &TbYangfenTransaction{})
}

// 新增交易记录
func (m *tbYangfenTransactionModel) Create(tx *TbYangfenTransaction) (int64, error) {
	return sqlite.DB.Insert(m.TableName(), tx)
}

// 按交易号查
func (m *tbYangfenTransactionModel) GetByTransactionId(transactionId string) (*TbYangfenTransaction, error) {
	var row TbYangfenTransaction
	err := sqlite.DB.FindOne(m.TableName(), &row, "transaction_id = ?", transactionId)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// 按用户查交易列表
func (m *tbYangfenTransactionModel) ListByUid(uid string, limit int) ([]*TbYangfenTransaction, error) {
	var transactions []*TbYangfenTransaction
	err := sqlite.DB.Find(m.TableName(), &transactions, "uid = ? ORDER BY id DESC LIMIT ?", uid, limit)
	return transactions, err
}

// 删除用户所有交易记录
func (m *tbYangfenTransactionModel) DeleteByUid(uid string) error {
	return sqlite.DB.DeleteByWhere(m.TableName(), "uid = ?", uid)
}
