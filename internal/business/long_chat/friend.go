package long_chat

import (
	"errors"
	"fmt"

	chatModel "github.com/armylong/armylong-go/internal/model/long_chat"
	userModel "github.com/armylong/armylong-go/internal/model/user"
)

type friendBusiness struct{}

var FriendBusiness = &friendBusiness{}

// AddFriend 添加好友
// uid=发起方, friendUid=被添加方
// 创建一条记录: (uid, friendUid, pending)
// 互加逻辑: 如果发现对方也加过我(friendUid→uid pending)，直接两条都变accepted
func (b *friendBusiness) AddFriend(uid, friendUid int64) error {
	if uid == friendUid {
		return errors.New("不能添加自己为好友")
	}

	// 检查自己是否已经加过对方
	existing, err := chatModel.TbLongChatFriendModel.GetByUidAndFriendUid(uid, friendUid)
	if err == nil && existing != nil {
		switch existing.Status {
		case chatModel.FriendStatusAccepted:
			return errors.New("已经是好友了")
		case chatModel.FriendStatusPending:
			return errors.New("已发送过好友请求，请等待对方确认")
		case chatModel.FriendStatusRejected:
			return errors.New("对方已拒绝您的好友请求")
		}
	}

	// 互加逻辑：对方也加过我（friendUid→uid pending），直接互加成功
	reverseExisting, err := chatModel.TbLongChatFriendModel.GetByUidAndFriendUid(friendUid, uid)
	if err == nil && reverseExisting != nil {
		if reverseExisting.Status == chatModel.FriendStatusPending {
			// 对方的记录变accepted
			if err := chatModel.TbLongChatFriendModel.Accept(friendUid, uid); err != nil {
				return err
			}
			// 补一条我的记录，直接accepted
			myRecord := &chatModel.TbLongChatFriend{
				Uid:       uid,
				FriendUid: friendUid,
				Status:    chatModel.FriendStatusAccepted,
			}
			if err := chatModel.TbLongChatFriendModel.Create(myRecord); err != nil {
				return fmt.Errorf("创建好友记录失败: %w", err)
			}
			return nil
		}
	}

	// 正常发起请求：一条记录 (uid, friendUid, pending)
	record := &chatModel.TbLongChatFriend{
		Uid:       uid,
		FriendUid: friendUid,
		Status:    chatModel.FriendStatusPending,
	}
	if err := chatModel.TbLongChatFriendModel.Create(record); err != nil {
		return fmt.Errorf("创建好友请求失败: %w", err)
	}

	return nil
}

// AcceptFriend 通过好友请求
// uid=被添加方(当前用户), friendUid=发起方
// 原记录变accepted，再补一条 (uid, friendUid, accepted)
func (b *friendBusiness) AcceptFriend(uid, friendUid int64) error {
	// 查对方发来的请求 (friendUid→uid)
	record, err := chatModel.TbLongChatFriendModel.GetByUidAndFriendUid(friendUid, uid)
	if err != nil || record == nil {
		return errors.New("好友请求不存在")
	}

	if record.Status != chatModel.FriendStatusPending {
		return errors.New("好友请求状态异常")
	}

	// 原记录变accepted
	if err := chatModel.TbLongChatFriendModel.Accept(friendUid, uid); err != nil {
		return err
	}

	// 补一条我的记录 (uid→friendUid, accepted)
	myRecord := &chatModel.TbLongChatFriend{
		Uid:       uid,
		FriendUid: friendUid,
		Status:    chatModel.FriendStatusAccepted,
	}
	if err := chatModel.TbLongChatFriendModel.Create(myRecord); err != nil {
		return fmt.Errorf("创建好友记录失败: %w", err)
	}

	return nil
}

// RejectFriend 拒绝好友请求
// uid=被添加方(当前用户), friendUid=发起方
// 只需把原记录变rejected
func (b *friendBusiness) RejectFriend(uid, friendUid int64) error {
	record, err := chatModel.TbLongChatFriendModel.GetByUidAndFriendUid(friendUid, uid)
	if err != nil || record == nil {
		return errors.New("好友请求不存在")
	}

	if record.Status != chatModel.FriendStatusPending {
		return errors.New("好友请求状态异常")
	}

	return chatModel.TbLongChatFriendModel.Reject(friendUid, uid)
}

// DeleteFriend 删除好友，双向删除
func (b *friendBusiness) DeleteFriend(uid, friendUid int64) error {
	_ = chatModel.TbLongChatFriendModel.Delete(uid, friendUid)
	_ = chatModel.TbLongChatFriendModel.Delete(friendUid, uid)
	return nil
}

// ListFriends 查询已通过的好友列表
// uid=我，查我发起且已通过的
func (b *friendBusiness) ListFriends(uid int64) ([]*chatModel.TbLongChatFriend, error) {
	return chatModel.TbLongChatFriendModel.ListFriends(uid)
}

// ListPendingRequests 查询待确认的好友请求
// 查别人加我且待确认的 (friend_uid=我, pending)
func (b *friendBusiness) ListPendingRequests(uid int64) ([]*chatModel.TbLongChatFriend, error) {
	return chatModel.TbLongChatFriendModel.ListPending(uid)
}

// IsFriend 判断两人是否为好友
func (b *friendBusiness) IsFriend(uid, friendUid int64) (bool, error) {
	return chatModel.TbLongChatFriendModel.IsFriend(uid, friendUid)
}

// SearchUser 搜索用户，返回用户列表及是否已是好友
func (b *friendBusiness) SearchUser(uid int64, keyword string) ([]*SearchUserItem, error) {
	users, err := userModel.TbUserModel.Search(keyword, 20)
	if err != nil {
		return nil, err
	}

	list := make([]*SearchUserItem, 0, len(users))
	for _, u := range users {
		if u.Uid == uid {
			continue
		}
		isFriend, _ := b.IsFriend(uid, u.Uid)
		u.ClearPassword()
		list = append(list, &SearchUserItem{
			Uid:      u.Uid,
			Account:  u.Account,
			Name:     u.Name,
			IsFriend: isFriend,
		})
	}
	return list, nil
}

// SearchUserItem 搜索用户结果项
type SearchUserItem struct {
	Uid      int64  `json:"uid"`
	Account  string `json:"account"`
	Name     string `json:"name"`
	IsFriend bool   `json:"is_friend"`
}
