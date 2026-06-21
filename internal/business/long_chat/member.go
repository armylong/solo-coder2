package long_chat

import (
	"errors"
	"fmt"

	chatModel "github.com/armylong/armylong-go/internal/model/long_chat"
	userModel "github.com/armylong/armylong-go/internal/model/user"
)

type memberBusiness struct{}

var MemberBusiness = &memberBusiness{}

// JoinChat 加入聊天，加 member + list 记录 + 系统消息
func (b *memberBusiness) JoinChat(chatId string, uid int64) error {
	chat, err := chatModel.TbLongChatModel.GetByChatId(chatId)
	if err != nil || chat == nil {
		return errors.New("聊天不存在")
	}

	if chat.Status == chatModel.ChatStatusDismiss {
		return errors.New("聊天已解散")
	}

	existing, err := chatModel.TbLongChatMemberModel.GetByChatIdAndUid(chatId, uid)
	if err == nil && existing != nil {
		return errors.New("您已在此聊天中")
	}

	if err := chatModel.TbLongChatMemberModel.Create(&chatModel.TbLongChatMember{
		ChatId: chatId,
		Uid:    uid,
	}); err != nil {
		return fmt.Errorf("加入聊天失败: %w", err)
	}

	chatName := chat.ChatName
	if chat.ChatType == chatModel.ChatTypePrivate {
		chatName = ""
	}

	if err := chatModel.TbLongChatListModel.Create(&chatModel.TbLongChatList{
		ChatId:   chatId,
		Uid:      uid,
		ChatName: chatName,
	}); err != nil {
		return fmt.Errorf("创建列表记录失败: %w", err)
	}

	user, _ := userModel.TbUserModel.GetByUid(uid)
	userName := "用户"
	if user != nil {
		userName = user.Name
	}
	systemMsg := fmt.Sprintf("%s加入了群聊", userName)
	_ = chatModel.TbLongChatMessageModel.CreateSystemMsg(chatId, systemMsg)

	return nil
}

// LeaveChat 退出聊天，删 member + list 记录 + 系统消息
func (b *memberBusiness) LeaveChat(chatId string, uid int64) error {
	chat, err := chatModel.TbLongChatModel.GetByChatId(chatId)
	if err != nil || chat == nil {
		return errors.New("聊天不存在")
	}

	if chat.ChatType == chatModel.ChatTypePrivate {
		return errors.New("单聊不能退出")
	}

	if chat.OwnerUid == uid {
		return errors.New("群主不能退出，请先转让群主或解散群聊")
	}

	member, err := chatModel.TbLongChatMemberModel.GetByChatIdAndUid(chatId, uid)
	if err != nil || member == nil {
		return errors.New("您不在此聊天中")
	}

	if err := chatModel.TbLongChatMemberModel.Delete(chatId, uid); err != nil {
		return fmt.Errorf("退出聊天失败: %w", err)
	}

	_ = chatModel.TbLongChatListModel.Delete(chatId, uid)

	user, _ := userModel.TbUserModel.GetByUid(uid)
	userName := "用户"
	if user != nil {
		userName = user.Name
	}
	systemMsg := fmt.Sprintf("%s退出了群聊", userName)
	_ = chatModel.TbLongChatMessageModel.CreateSystemMsg(chatId, systemMsg)

	return nil
}

// KickMember 踢人，校验群主权限
func (b *memberBusiness) KickMember(chatId string, operatorUid, targetUid int64) error {
	chat, err := chatModel.TbLongChatModel.GetByChatId(chatId)
	if err != nil || chat == nil {
		return errors.New("聊天不存在")
	}

	if chat.OwnerUid != operatorUid {
		return errors.New("只有群主才能踢人")
	}

	if operatorUid == targetUid {
		return errors.New("不能踢自己")
	}

	member, err := chatModel.TbLongChatMemberModel.GetByChatIdAndUid(chatId, targetUid)
	if err != nil || member == nil {
		return errors.New("该用户不在此聊天中")
	}

	if err := chatModel.TbLongChatMemberModel.Delete(chatId, targetUid); err != nil {
		return fmt.Errorf("踢出成员失败: %w", err)
	}

	_ = chatModel.TbLongChatListModel.Delete(chatId, targetUid)

	targetUser, _ := userModel.TbUserModel.GetByUid(targetUid)
	targetName := "用户"
	if targetUser != nil {
		targetName = targetUser.Name
	}
	systemMsg := fmt.Sprintf("%s被移出了群聊", targetName)
	_ = chatModel.TbLongChatMessageModel.CreateSystemMsg(chatId, systemMsg)

	return nil
}

// ListMembers 成员列表
func (b *memberBusiness) ListMembers(chatId string) ([]*chatModel.TbLongChatMember, error) {
	return chatModel.TbLongChatMemberModel.ListByChatId(chatId)
}
