package long_chat

import (
	"time"

	"github.com/armylong/go-library/service/sqlite"
)

// TbLongChatMember 聊天成员，只存当前成员，离开即删除
// 加群/退群历史通过系统消息追溯
type TbLongChatMember struct {
	ChatId   string    `json:"chat_id"` // 关联聊天ID
	Uid      int64     `json:"uid"`     // 用户ID
	JoinedAt time.Time `json:"joined_at"`
}

type tbLongChatMemberModel struct{}

var TbLongChatMemberModel = &tbLongChatMemberModel{}

func init() {
	_ = TbLongChatMemberModel.CreateTable()
}

func (m *tbLongChatMemberModel) TableName() string {
	return "tb_long_chat_member"
}

func (m *tbLongChatMemberModel) CreateTable() error {
	sql := `
	CREATE TABLE IF NOT EXISTS tb_long_chat_member (
		chat_id TEXT NOT NULL,
		uid INTEGER NOT NULL,
		joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (chat_id, uid)
	)
	`
	_, err := sqlite.DB.DB().Exec(sql)
	if err != nil {
		return err
	}
	return sqlite.DB.AutoMigrate(m.TableName(), &TbLongChatMember{})
}

// Create 添加单个成员
func (m *tbLongChatMemberModel) Create(member *TbLongChatMember) error {
	member.JoinedAt = time.Now()
	_, err := sqlite.DB.Insert(m.TableName(), member)
	return err
}

// BatchCreate 批量添加成员，用于建群时一次性加入所有成员
func (m *tbLongChatMemberModel) BatchCreate(chatId string, uids []int64) error {
	members := make([]*TbLongChatMember, 0, len(uids))
	now := time.Now()
	for _, uid := range uids {
		members = append(members, &TbLongChatMember{
			ChatId:   chatId,
			Uid:      uid,
			JoinedAt: now,
		})
	}
	return sqlite.DB.InsertBatch(m.TableName(), members)
}

// GetByChatIdAndUid 查询用户是否在聊天中
func (m *tbLongChatMemberModel) GetByChatIdAndUid(chatId string, uid int64) (*TbLongChatMember, error) {
	var member TbLongChatMember
	err := sqlite.DB.FindOne(m.TableName(), &member, "chat_id = ? AND uid = ?", chatId, uid)
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// ListByChatId 查询聊天的所有成员
func (m *tbLongChatMemberModel) ListByChatId(chatId string) ([]*TbLongChatMember, error) {
	var members []*TbLongChatMember
	err := sqlite.DB.Find(m.TableName(), &members, "chat_id = ?", chatId)
	return members, err
}

// ListUidsByChatId 查询聊天的所有成员uid列表，用于群发消息
func (m *tbLongChatMemberModel) ListUidsByChatId(chatId string) ([]int64, error) {
	members, err := m.ListByChatId(chatId)
	if err != nil {
		return nil, err
	}
	uids := make([]int64, 0, len(members))
	for _, member := range members {
		uids = append(uids, member.Uid)
	}
	return uids, nil
}

// ListChatIdsByUid 查询用户参与的所有聊天ID
func (m *tbLongChatMemberModel) ListChatIdsByUid(uid int64) ([]string, error) {
	var members []*TbLongChatMember
	err := sqlite.DB.Find(m.TableName(), &members, "uid = ?", uid)
	if err != nil {
		return nil, err
	}
	chatIds := make([]string, 0, len(members))
	for _, member := range members {
		chatIds = append(chatIds, member.ChatId)
	}
	return chatIds, nil
}

// Delete 移除成员（退群/被踢），直接删除记录
func (m *tbLongChatMemberModel) Delete(chatId string, uid int64) error {
	return sqlite.DB.DeleteByWhere(m.TableName(), "chat_id = ? AND uid = ?", chatId, uid)
}

// CountByChatId 统计聊天成员数
func (m *tbLongChatMemberModel) CountByChatId(chatId string) (int64, error) {
	return sqlite.DB.Count(m.TableName(), "chat_id = ?", chatId)
}
