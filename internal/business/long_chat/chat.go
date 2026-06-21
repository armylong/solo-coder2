package long_chat

import (
	"errors"
	"fmt"

	chatModel "github.com/armylong/armylong-go/internal/model/long_chat"
	userModel "github.com/armylong/armylong-go/internal/model/user"
)

type chatBusiness struct{}

var ChatBusiness = &chatBusiness{}

// CreatePrivateChat 创建单聊，同时创建双方 member 和 list 记录
func (b *chatBusiness) CreatePrivateChat(uid, targetUid int64) (string, error) {
	if uid == targetUid {
		return "", errors.New("不能和自己聊天")
	}

	targetUser, err := userModel.TbUserModel.GetByUid(targetUid)
	if err != nil || targetUser == nil {
		return "", errors.New("对方用户不存在")
	}

	currentUser, err := userModel.TbUserModel.GetByUid(uid)
	if err != nil || currentUser == nil {
		return "", errors.New("用户不存在")
	}

	chat := &chatModel.TbLongChat{
		ChatType: chatModel.ChatTypePrivate,
		OwnerUid: uid,
	}
	if err := chatModel.TbLongChatModel.Create(chat); err != nil {
		return "", fmt.Errorf("创建单聊失败: %w", err)
	}

	if err := chatModel.TbLongChatMemberModel.BatchCreate(chat.ChatId, []int64{uid, targetUid}); err != nil {
		return "", fmt.Errorf("创建成员记录失败: %w", err)
	}

	listItems := []*chatModel.TbLongChatList{
		{ChatId: chat.ChatId, Uid: uid, ChatName: targetUser.Name},
		{ChatId: chat.ChatId, Uid: targetUid, ChatName: currentUser.Name},
	}
	if err := chatModel.TbLongChatListModel.BatchCreate(listItems); err != nil {
		return "", fmt.Errorf("创建列表记录失败: %w", err)
	}

	return chat.ChatId, nil
}

// CreateGroup 创建群聊，同时创建所有 member 和 list 记录
func (b *chatBusiness) CreateGroup(ownerUid int64, groupName string, memberUids []int64) (string, error) {
	if groupName == "" {
		return "", errors.New("群名不能为空")
	}
	if len(memberUids) == 0 {
		return "", errors.New("群成员不能为空")
	}

	allUids := append([]int64{ownerUid}, memberUids...)

	chat := &chatModel.TbLongChat{
		ChatType: chatModel.ChatTypeGroup,
		ChatName: groupName,
		OwnerUid: ownerUid,
	}
	if err := chatModel.TbLongChatModel.Create(chat); err != nil {
		return "", fmt.Errorf("创建群聊失败: %w", err)
	}

	if err := chatModel.TbLongChatMemberModel.BatchCreate(chat.ChatId, allUids); err != nil {
		return "", fmt.Errorf("创建成员记录失败: %w", err)
	}

	listItems := make([]*chatModel.TbLongChatList, 0, len(allUids))
	for _, uid := range allUids {
		listItems = append(listItems, &chatModel.TbLongChatList{
			ChatId:   chat.ChatId,
			Uid:      uid,
			ChatName: groupName,
		})
	}
	if err := chatModel.TbLongChatListModel.BatchCreate(listItems); err != nil {
		return "", fmt.Errorf("创建列表记录失败: %w", err)
	}

	return chat.ChatId, nil
}

// GetChat 获取聊天信息
func (b *chatBusiness) GetChat(chatId string) (*chatModel.TbLongChat, error) {
	return chatModel.TbLongChatModel.GetByChatId(chatId)
}

// DismissChat 解散聊天，校验是否为群主
func (b *chatBusiness) DismissChat(chatId string, operatorUid int64) error {
	chat, err := chatModel.TbLongChatModel.GetByChatId(chatId)
	if err != nil || chat == nil {
		return errors.New("聊天不存在")
	}

	if chat.ChatType == chatModel.ChatTypePrivate {
		return errors.New("单聊不能解散")
	}

	if chat.OwnerUid != operatorUid {
		return errors.New("只有群主才能解散群聊")
	}

	if chat.Status == chatModel.ChatStatusDismiss {
		return errors.New("群聊已解散")
	}

	return chatModel.TbLongChatModel.Dismiss(chatId)
}

// TransferOwner 转让群主，校验当前是否为群主
func (b *chatBusiness) TransferOwner(chatId string, operatorUid, newOwnerUid int64) error {
	chat, err := chatModel.TbLongChatModel.GetByChatId(chatId)
	if err != nil || chat == nil {
		return errors.New("聊天不存在")
	}

	if chat.OwnerUid != operatorUid {
		return errors.New("只有群主才能转让")
	}

	if operatorUid == newOwnerUid {
		return errors.New("不能转让给自己")
	}

	isMember, err := chatModel.TbLongChatMemberModel.GetByChatIdAndUid(chatId, newOwnerUid)
	if err != nil || isMember == nil {
		return errors.New("目标用户不在群聊中")
	}

	chat.OwnerUid = newOwnerUid
	return chatModel.TbLongChatModel.Update(chat)
}

// GetOrCreatePrivateChat 获取或创建单聊（发消息时调用，避免重复创建）
func (b *chatBusiness) GetOrCreatePrivateChat(uid, targetUid int64) (string, error) {
	chatIds, err := chatModel.TbLongChatMemberModel.ListChatIdsByUid(uid)
	if err != nil {
		return "", fmt.Errorf("查询聊天列表失败: %w", err)
	}

	for _, chatId := range chatIds {
		chat, err := chatModel.TbLongChatModel.GetByChatId(chatId)
		if err != nil || chat == nil {
			continue
		}
		if chat.ChatType != chatModel.ChatTypePrivate {
			continue
		}
		member, err := chatModel.TbLongChatMemberModel.GetByChatIdAndUid(chatId, targetUid)
		if err == nil && member != nil {
			return chatId, nil
		}
	}

	return b.CreatePrivateChat(uid, targetUid)
}
