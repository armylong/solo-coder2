package long_chat

import (
	"time"

	"github.com/armylong/go-library/service/sqlite"
)

// TbLongChatList 用户聊天列表，每个用户每个会话一条记录
// 冗余chat_name避免查询时JOIN，聊天列表一条SQL搞定
type TbLongChatList struct {
	ChatId    string    `json:"chat_id"`   // 关联聊天ID
	Uid       int64     `json:"uid"`       // 这条记录属于谁
	ChatName  string    `json:"chat_name"` // 单聊存对方昵称, 群聊存群名
	Unread    int       `json:"unread"`    // 未读数
	LastMsg   string    `json:"last_msg"`  // 最新消息摘要
	LastMsgAt time.Time `json:"last_msg_at"`
	CreatedAt time.Time `json:"created_at"`
}

type tbLongChatListModel struct{}

var TbLongChatListModel = &tbLongChatListModel{}

func init() {
	_ = TbLongChatListModel.CreateTable()
}

func (m *tbLongChatListModel) TableName() string {
	return "tb_long_chat_list"
}

func (m *tbLongChatListModel) CreateTable() error {
	sql := `
	CREATE TABLE IF NOT EXISTS tb_long_chat_list (
		chat_id TEXT NOT NULL,
		uid INTEGER NOT NULL,
		chat_name TEXT,
		unread INTEGER DEFAULT 0,
		last_msg TEXT,
		last_msg_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (chat_id, uid)
	)
	`
	_, err := sqlite.DB.DB().Exec(sql)
	if err != nil {
		return err
	}
	return sqlite.DB.AutoMigrate(m.TableName(), &TbLongChatList{})
}

// Create 创建单条列表记录
func (m *tbLongChatListModel) Create(item *TbLongChatList) error {
	item.CreatedAt = time.Now()
	_, err := sqlite.DB.Insert(m.TableName(), item)
	return err
}

// BatchCreate 批量创建列表记录，用于建群/发起单聊时为每个参与者创建
func (m *tbLongChatListModel) BatchCreate(items []*TbLongChatList) error {
	return sqlite.DB.InsertBatch(m.TableName(), items)
}

// GetByChatIdAndUid 查询用户在某会话的列表记录
func (m *tbLongChatListModel) GetByChatIdAndUid(chatId string, uid int64) (*TbLongChatList, error) {
	var item TbLongChatList
	err := sqlite.DB.FindOne(m.TableName(), &item, "chat_id = ? AND uid = ?", chatId, uid)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// ListByUid 查询用户的聊天列表，按最新消息时间倒序
func (m *tbLongChatListModel) ListByUid(uid int64) ([]*TbLongChatList, error) {
	var items []*TbLongChatList
	err := sqlite.DB.Find(m.TableName(), &items, "uid = ? ORDER BY last_msg_at DESC", uid)
	return items, err
}

// UpdateLastMsg 更新最新消息摘要和时间
func (m *tbLongChatListModel) UpdateLastMsg(chatId string, uid int64, lastMsg string, lastMsgAt time.Time) error {
	return sqlite.DB.UpdateByWhere(m.TableName(),
		&TbLongChatList{LastMsg: lastMsg, LastMsgAt: lastMsgAt},
		[]string{"last_msg", "last_msg_at"},
		"chat_id = ? AND uid = ?", []any{chatId, uid},
	)
}

// IncrUnread 单个用户未读数+1
func (m *tbLongChatListModel) IncrUnread(chatId string, uid int64) error {
	_, err := sqlite.DB.DB().Exec(
		"UPDATE tb_long_chat_list SET unread = unread + 1 WHERE chat_id = ? AND uid = ?",
		chatId, uid,
	)
	return err
}

// BatchIncrUnread 批量未读数+1，排除发送者，用于群发消息时更新所有接收者的未读数
func (m *tbLongChatListModel) BatchIncrUnread(chatId string, excludeUid int64) error {
	_, err := sqlite.DB.DB().Exec(
		"UPDATE tb_long_chat_list SET unread = unread + 1 WHERE chat_id = ? AND uid != ?",
		chatId, excludeUid,
	)
	return err
}

// ClearUnread 清零未读数，用户打开聊天详情时调用
func (m *tbLongChatListModel) ClearUnread(chatId string, uid int64) error {
	return sqlite.DB.UpdateByWhere(m.TableName(),
		&TbLongChatList{Unread: 0},
		[]string{"unread"},
		"chat_id = ? AND uid = ?", []any{chatId, uid},
	)
}

// UpdateChatName 更新聊天名称，用于对方改名或群名修改时同步
func (m *tbLongChatListModel) UpdateChatName(chatId string, uid int64, chatName string) error {
	return sqlite.DB.UpdateByWhere(m.TableName(),
		&TbLongChatList{ChatName: chatName},
		[]string{"chat_name"},
		"chat_id = ? AND uid = ?", []any{chatId, uid},
	)
}

// Delete 删除列表记录
func (m *tbLongChatListModel) Delete(chatId string, uid int64) error {
	return sqlite.DB.DeleteByWhere(m.TableName(), "chat_id = ? AND uid = ?", chatId, uid)
}
