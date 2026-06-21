package long_chat

import (
	"context"

	longChatBiz "github.com/armylong/armylong-go/internal/business/long_chat"
	"github.com/armylong/armylong-go/internal/middlewares"
	longChatCs "github.com/armylong/armylong-go/internal/cs/long_chat"
	userModel "github.com/armylong/armylong-go/internal/model/user"
)

// MessageController 消息管理
type MessageController struct{}

// 发送消息
func (c *MessageController) ActionSendMessage(ctx context.Context, req *longChatCs.SendMessageRequest) (*longChatCs.SendMessageResponse, error) {
	uid := middlewares.GetLoginUIDFromContext(ctx)
	msg, err := longChatBiz.MessageBusiness.SendMessage(req.ChatId, uid, req.MsgType, req.Content)
	if err != nil {
		return nil, err
	}
	return &longChatCs.SendMessageResponse{Msg: msg}, nil
}

// 历史消息
func (c *MessageController) ActionListMessages(ctx context.Context, req *longChatCs.ListMessagesRequest) (*longChatCs.ListMessagesResponse, error) {
	uid := middlewares.GetLoginUIDFromContext(ctx)
	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	msgs, err := longChatBiz.MessageBusiness.ListMessages(req.ChatId, uid, limit, req.Offset)
	if err != nil {
		return nil, err
	}

	// 批量查发送者昵称，避免N+1
	uidNameMap := make(map[int64]string)
	for _, m := range msgs {
		uidNameMap[m.FromUid] = ""
	}
	for u := range uidNameMap {
		user, err := userModel.TbUserModel.GetByUid(u)
		if err == nil && user != nil {
			uidNameMap[u] = user.Name
		}
	}

	list := make([]*longChatCs.MessageItem, 0, len(msgs))
	for _, m := range msgs {
		list = append(list, &longChatCs.MessageItem{
			MsgId:      m.MsgId,
			ChatId:     m.ChatId,
			FromUid:    m.FromUid,
			SenderName: uidNameMap[m.FromUid],
			MsgType:    m.MsgType,
			Content:    m.Content,
			CreatedAt:  m.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return &longChatCs.ListMessagesResponse{Uid: uid, List: list}, nil
}

// 清零未读
func (c *MessageController) ActionClearUnread(ctx context.Context, req *longChatCs.ClearUnreadRequest) (*longChatCs.ClearUnreadResponse, error) {
	uid := middlewares.GetLoginUIDFromContext(ctx)
	if err := longChatBiz.MessageBusiness.ClearUnread(req.ChatId, uid); err != nil {
		return nil, err
	}
	return &longChatCs.ClearUnreadResponse{}, nil
}
