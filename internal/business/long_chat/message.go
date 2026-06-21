package long_chat

import (
	"errors"
	"fmt"
	"time"

	chatModel "github.com/armylong/armylong-go/internal/model/long_chat"
	userModel "github.com/armylong/armylong-go/internal/model/user"
	ws "github.com/armylong/armylong-go/internal/websocket"
	libWs "github.com/armylong/go-library/service/websocket"
)

type messageBusiness struct{}

var MessageBusiness = &messageBusiness{}

// SendMessage 发送消息
// 内部流程：校验 → 插入消息 → 更新列表 → 推送WebSocket
func (b *messageBusiness) SendMessage(chatId string, fromUid int64, msgType string, content string) (*chatModel.TbLongChatMessage, error) {
	chat, err := chatModel.TbLongChatModel.GetByChatId(chatId)
	if err != nil || chat == nil {
		return nil, errors.New("聊天不存在")
	}

	if chat.Status == chatModel.ChatStatusDismiss {
		return nil, errors.New("聊天已解散")
	}

	member, err := chatModel.TbLongChatMemberModel.GetByChatIdAndUid(chatId, fromUid)
	if err != nil || member == nil {
		return nil, errors.New("您不在此聊天中")
	}

	msg := &chatModel.TbLongChatMessage{
		ChatId:  chatId,
		FromUid: fromUid,
		MsgType: msgType,
		Content: content,
	}
	if err := chatModel.TbLongChatMessageModel.Create(msg); err != nil {
		return nil, fmt.Errorf("发送消息失败: %w", err)
	}

	if err := b.updateChatList(chatId, fromUid, msg); err != nil {
		return nil, fmt.Errorf("更新聊天列表失败: %w", err)
	}

	b.pushToMembers(chatId, fromUid, msg)

	return msg, nil
}

// ListMessages 分页获取聊天记录，按时间倒序
func (b *messageBusiness) ListMessages(chatId string, uid int64, limit, offset int) ([]*chatModel.TbLongChatMessage, error) {
	member, err := chatModel.TbLongChatMemberModel.GetByChatIdAndUid(chatId, uid)
	if err != nil || member == nil {
		return nil, errors.New("您不在此聊天中")
	}

	return chatModel.TbLongChatMessageModel.ListByChatId(chatId, limit, offset)
}

// ClearUnread 清零未读数，打开聊天详情时调用
func (b *messageBusiness) ClearUnread(chatId string, uid int64) error {
	return chatModel.TbLongChatListModel.ClearUnread(chatId, uid)
}

// updateChatList 更新所有参与者的聊天列表（last_msg、unread）
func (b *messageBusiness) updateChatList(chatId string, fromUid int64, msg *chatModel.TbLongChatMessage) error {
	lastMsg := b.buildLastMsg(msg)
	now := msg.CreatedAt

	members, err := chatModel.TbLongChatMemberModel.ListByChatId(chatId)
	if err != nil {
		return err
	}

	for _, m := range members {
		if err := chatModel.TbLongChatListModel.UpdateLastMsg(chatId, m.Uid, lastMsg, now); err != nil {
			return err
		}
		if m.Uid != fromUid {
			if err := chatModel.TbLongChatListModel.IncrUnread(chatId, m.Uid); err != nil {
				return err
			}
		}
	}

	return nil
}

// buildLastMsg 构建聊天列表的最新消息摘要
func (b *messageBusiness) buildLastMsg(msg *chatModel.TbLongChatMessage) string {
	switch msg.MsgType {
	case chatModel.MsgTypeText:
		return msg.Content
	case chatModel.MsgTypeImage:
		return "[图片]"
	case chatModel.MsgTypeVoice:
		return "[语音]"
	case chatModel.MsgTypeSystem:
		return msg.Content
	default:
		return "[消息]"
	}
}

// pushToMembers 通过WebSocket推送消息给在线成员，统一走 PushToGroup 频道推送
func (b *messageBusiness) pushToMembers(chatId string, fromUid int64, msg *chatModel.TbLongChatMessage) {
	if ws.Manager == nil {
		return
	}

	// 查发送者昵称
	senderName := ""
	u, err := userModel.TbUserModel.GetByUid(fromUid)
	if err == nil && u != nil {
		senderName = u.Name
	}

	pushData := map[string]any{
		"chat_id":     msg.ChatId,
		"msg_id":      msg.MsgId,
		"from_uid":    msg.FromUid,
		"sender_name": senderName,
		"msg_type":    msg.MsgType,
		"content":     msg.Content,
		"created_at":  msg.CreatedAt.Format(time.RFC3339),
	}
	wsMsg := libWs.NewMessage(ws.TypeLongChat, pushData)
	ws.Manager.PushToGroup(chatId, wsMsg)
}
