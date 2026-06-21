package long_chat

import (
	"fmt"
	"time"

	"github.com/armylong/go-library/service/snowflake"
	"github.com/armylong/go-library/service/sqlite"
)

// 消息类型
const (
	MsgTypeText   = "text"   // 文本
	MsgTypeImage  = "image"  // 图片
	MsgTypeVoice  = "voice"  // 语音
	MsgTypeSystem = "system" // 系统消息（加群/退群提示等）
)

// TbLongChatMessage 消息
// 文本消息content为纯文本，非文本消息content为JSON
type TbLongChatMessage struct {
	MsgId     string    `json:"msg_id" db:"pk"` // 雪花ID
	ChatId    string    `json:"chat_id"`        // 关联聊天ID
	FromUid   int64     `json:"from_uid"`       // 发送者uid, 系统消息时为0
	MsgType   string    `json:"msg_type"`       // 消息类型: text/image/voice/system/...
	Content   string    `json:"content"`        // 文本时为纯文本, 非文本时为JSON
	CreatedAt time.Time `json:"created_at"`
}

type tbLongChatMessageModel struct{}

var TbLongChatMessageModel = &tbLongChatMessageModel{}

func init() {
	_ = TbLongChatMessageModel.CreateTable()
}

func (m *tbLongChatMessageModel) TableName() string {
	return "tb_long_chat_message"
}

func (m *tbLongChatMessageModel) CreateTable() error {
	sql := `
	CREATE TABLE IF NOT EXISTS tb_long_chat_message (
		msg_id TEXT PRIMARY KEY,
		chat_id TEXT NOT NULL,
		from_uid INTEGER NOT NULL,
		msg_type TEXT DEFAULT 'text',
		content TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)
	`
	_, err := sqlite.DB.DB().Exec(sql)
	if err != nil {
		return err
	}
	return sqlite.DB.AutoMigrate(m.TableName(), &TbLongChatMessage{})
}

// Create 发送消息，自动生成雪花ID
func (m *tbLongChatMessageModel) Create(msg *TbLongChatMessage) error {
	msg.MsgId = fmt.Sprintf("%d", snowflake.Generate())
	msg.CreatedAt = time.Now()
	_, err := sqlite.DB.Insert(m.TableName(), msg)
	return err
}

// GetByMsgId 按msg_id查询
func (m *tbLongChatMessageModel) GetByMsgId(msgId string) (*TbLongChatMessage, error) {
	var msg TbLongChatMessage
	msg.MsgId = msgId
	err := sqlite.DB.GetByPkId(m.TableName(), &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

// ListByChatId 分页查询聊天记录，按时间倒序
func (m *tbLongChatMessageModel) ListByChatId(chatId string, limit, offset int) ([]*TbLongChatMessage, error) {
	var msgs []*TbLongChatMessage
	err := sqlite.DB.Find(m.TableName(), &msgs, "chat_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?", chatId, limit, offset)
	return msgs, err
}

// CountByChatId 统计聊天消息数
func (m *tbLongChatMessageModel) CountByChatId(chatId string) (int64, error) {
	return sqlite.DB.Count(m.TableName(), "chat_id = ?", chatId)
}

// CreateSystemMsg 发送系统消息，from_uid为0，用于加群/退群提示等
func (m *tbLongChatMessageModel) CreateSystemMsg(chatId string, content string) error {
	return m.Create(&TbLongChatMessage{
		ChatId:  chatId,
		FromUid: 0,
		MsgType: MsgTypeSystem,
		Content: content,
	})
}
