package long_chat

import (
	"time"

	"github.com/armylong/go-library/service/sqlite"
)

// 好友状态
const (
	FriendStatusPending  = 1 // 待确认
	FriendStatusAccepted = 2 // 已通过
	FriendStatusRejected = 3 // 已拒绝
)

// TbLongChatFriend 好友关系
// uid=发起方, friend_uid=被添加方
// A加B时创建一条记录: (A的uid, B的uid, pending)
// B同意后: 原记录变accepted，再补一条 (B的uid, A的uid, accepted)
// 互加逻辑: A加B时如果发现B也加过A(pending)，直接两条都变accepted
type TbLongChatFriend struct {
	Uid       int64     `json:"uid"`        // 发起方uid
	FriendUid int64     `json:"friend_uid"` // 被添加方uid
	Status    int       `json:"status"`     // 状态: 1-待确认 2-已通过 3-已拒绝
	CreatedAt time.Time `json:"created_at"`
}

type tbLongChatFriendModel struct{}

var TbLongChatFriendModel = &tbLongChatFriendModel{}

func init() {
	_ = TbLongChatFriendModel.CreateTable()
}

func (m *tbLongChatFriendModel) TableName() string {
	return "tb_long_chat_friend"
}

func (m *tbLongChatFriendModel) CreateTable() error {
	sql := `
	CREATE TABLE IF NOT EXISTS tb_long_chat_friend (
		uid INTEGER NOT NULL,
		friend_uid INTEGER NOT NULL,
		status INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (uid, friend_uid)
	)
	`
	_, err := sqlite.DB.DB().Exec(sql)
	if err != nil {
		return err
	}
	return sqlite.DB.AutoMigrate(m.TableName(), &TbLongChatFriend{})
}

// Create 创建好友关系记录
func (m *tbLongChatFriendModel) Create(friend *TbLongChatFriend) error {
	friend.CreatedAt = time.Now()
	_, err := sqlite.DB.Insert(m.TableName(), friend)
	return err
}

// GetByUidAndFriendUid 查询两人之间的好友关系（uid=发起方, friend_uid=被添加方）
func (m *tbLongChatFriendModel) GetByUidAndFriendUid(uid, friendUid int64) (*TbLongChatFriend, error) {
	var friend TbLongChatFriend
	err := sqlite.DB.FindOne(m.TableName(), &friend, "uid = ? AND friend_uid = ?", uid, friendUid)
	if err != nil {
		return nil, err
	}
	return &friend, nil
}

// ListFriends 查询已通过的好友列表（uid=我，即我发起且已通过的）
func (m *tbLongChatFriendModel) ListFriends(uid int64) ([]*TbLongChatFriend, error) {
	var friends []*TbLongChatFriend
	err := sqlite.DB.Find(m.TableName(), &friends, "uid = ? AND status = ?", uid, FriendStatusAccepted)
	return friends, err
}

// ListPending 查询待确认的好友请求（friend_uid=我，即别人加我待确认的）
func (m *tbLongChatFriendModel) ListPending(friendUid int64) ([]*TbLongChatFriend, error) {
	var friends []*TbLongChatFriend
	err := sqlite.DB.Find(m.TableName(), &friends, "friend_uid = ? AND status = ?", friendUid, FriendStatusPending)
	return friends, err
}

// Accept 通过好友请求
func (m *tbLongChatFriendModel) Accept(uid, friendUid int64) error {
	return sqlite.DB.UpdateByWhere(m.TableName(),
		&TbLongChatFriend{Status: FriendStatusAccepted},
		[]string{"status"},
		"uid = ? AND friend_uid = ?", []any{uid, friendUid},
	)
}

// Reject 拒绝好友请求
func (m *tbLongChatFriendModel) Reject(uid, friendUid int64) error {
	return sqlite.DB.UpdateByWhere(m.TableName(),
		&TbLongChatFriend{Status: FriendStatusRejected},
		[]string{"status"},
		"uid = ? AND friend_uid = ?", []any{uid, friendUid},
	)
}

// Delete 删除好友关系
func (m *tbLongChatFriendModel) Delete(uid, friendUid int64) error {
	return sqlite.DB.DeleteByWhere(m.TableName(), "uid = ? AND friend_uid = ?", uid, friendUid)
}

// IsFriend 判断两人是否为好友（任一方向有accepted记录即可）
func (m *tbLongChatFriendModel) IsFriend(uid, friendUid int64) (bool, error) {
	count, err := sqlite.DB.Count(m.TableName(), "(uid = ? AND friend_uid = ? OR uid = ? AND friend_uid = ?) AND status = ?",
		uid, friendUid, friendUid, uid, FriendStatusAccepted)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
