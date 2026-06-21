package long_chat

import (
	"fmt"
	"time"

	"github.com/armylong/go-library/service/snowflake"
	"github.com/armylong/go-library/service/sqlite"
)

// 聊天类型
const (
	ChatTypePrivate = 1 // 单聊
	ChatTypeGroup   = 2 // 群聊
)

// 聊天状态
const (
	ChatStatusNormal  = 1 // 正常
	ChatStatusDismiss = 2 // 已解散
)

// TbLongChat 聊天实体，单聊和群聊统一
type TbLongChat struct {
	ChatId    string    `json:"chat_id" db:"pk"` // 雪花ID
	ChatType  int       `json:"chat_type"`       // 聊天类型: 1-单聊 2-群聊
	ChatName  string    `json:"chat_name"`       // 群聊存群名, 单聊可为空
	OwnerUid  int64     `json:"owner_uid"`       // 群主uid, 单聊时为发起人uid
	Status    int       `json:"status"`          // 状态: 1-正常 2-已解散
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type tbLongChatModel struct{}

var TbLongChatModel = &tbLongChatModel{}

func init() {
	_ = TbLongChatModel.CreateTable()
}

func (m *tbLongChatModel) TableName() string {
	return "tb_long_chat"
}

func (m *tbLongChatModel) CreateTable() error {
	sql := `
	CREATE TABLE IF NOT EXISTS tb_long_chat (
		chat_id TEXT PRIMARY KEY,
		chat_type INTEGER DEFAULT 1,
		chat_name TEXT,
		owner_uid INTEGER DEFAULT 0,
		status INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err := sqlite.DB.DB().Exec(sql)
	if err != nil {
		return err
	}
	return sqlite.DB.AutoMigrate(m.TableName(), &TbLongChat{})
}

// Create 创建聊天，自动生成雪花ID
func (m *tbLongChatModel) Create(chat *TbLongChat) error {
	chat.ChatId = fmt.Sprintf("%d", snowflake.Generate())
	chat.CreatedAt = time.Now()
	chat.UpdatedAt = time.Now()
	_, err := sqlite.DB.Insert(m.TableName(), chat)
	return err
}

// GetByChatId 按chat_id查询
func (m *tbLongChatModel) GetByChatId(chatId string) (*TbLongChat, error) {
	var chat TbLongChat
	chat.ChatId = chatId
	err := sqlite.DB.GetByPkId(m.TableName(), &chat)
	if err != nil {
		return nil, err
	}
	return &chat, nil
}

// Update 更新聊天信息
func (m *tbLongChatModel) Update(chat *TbLongChat) error {
	chat.UpdatedAt = time.Now()
	return sqlite.DB.UpdateByPkId(m.TableName(), chat)
}

// Dismiss 解散聊天，状态改为已解散
func (m *tbLongChatModel) Dismiss(chatId string) error {
	chat, err := m.GetByChatId(chatId)
	if err != nil {
		return err
	}
	chat.Status = ChatStatusDismiss
	chat.UpdatedAt = time.Now()
	return sqlite.DB.UpdateByPkId(m.TableName(), chat)
}
