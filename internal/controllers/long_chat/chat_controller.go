package long_chat

import (
	"context"

	longChatBiz "github.com/armylong/armylong-go/internal/business/long_chat"
	chatModel "github.com/armylong/armylong-go/internal/model/long_chat"
	"github.com/armylong/armylong-go/internal/middlewares"
	longChatCs "github.com/armylong/armylong-go/internal/cs/long_chat"
)

// ChatController 聊天管理
type ChatController struct{}

// 聊天列表
func (c *ChatController) ActionGetChatList(ctx context.Context, req *longChatCs.GetChatListRequest) (*longChatCs.GetChatListResponse, error) {
	uid := middlewares.GetLoginUIDFromContext(ctx)
	items, err := chatModel.TbLongChatListModel.ListByUid(uid)
	if err != nil {
		return nil, err
	}

	list := make([]*longChatCs.ChatListItem, 0, len(items))
	for _, item := range items {
		chat, _ := chatModel.TbLongChatModel.GetByChatId(item.ChatId)
		chatType := 0
		if chat != nil {
			chatType = chat.ChatType
		}
		lastMsgAt := ""
		if !item.LastMsgAt.IsZero() {
			lastMsgAt = item.LastMsgAt.Format("2006-01-02 15:04:05")
		}
		list = append(list, &longChatCs.ChatListItem{
			ChatId:    item.ChatId,
			ChatType:  chatType,
			ChatName:  item.ChatName,
			Unread:    item.Unread,
			LastMsg:   item.LastMsg,
			LastMsgAt: lastMsgAt,
		})
	}

	return &longChatCs.GetChatListResponse{List: list}, nil
}

// 创建单聊
func (c *ChatController) ActionCreatePrivateChat(ctx context.Context, req *longChatCs.CreatePrivateChatRequest) (*longChatCs.CreatePrivateChatResponse, error) {
	uid := middlewares.GetLoginUIDFromContext(ctx)
	chatId, err := longChatBiz.ChatBusiness.GetOrCreatePrivateChat(uid, req.TargetUid)
	if err != nil {
		return nil, err
	}
	return &longChatCs.CreatePrivateChatResponse{ChatId: chatId}, nil
}

// 创建群聊
func (c *ChatController) ActionCreateGroup(ctx context.Context, req *longChatCs.CreateGroupRequest) (*longChatCs.CreateGroupResponse, error) {
	uid := middlewares.GetLoginUIDFromContext(ctx)
	chatId, err := longChatBiz.ChatBusiness.CreateGroup(uid, req.GroupName, req.MemberUids)
	if err != nil {
		return nil, err
	}
	return &longChatCs.CreateGroupResponse{ChatId: chatId}, nil
}

// 聊天详情
func (c *ChatController) ActionGetChatDetail(ctx context.Context, req *longChatCs.GetChatDetailRequest) (*longChatCs.GetChatDetailResponse, error) {
	chat, err := longChatBiz.ChatBusiness.GetChat(req.ChatId)
	if err != nil {
		return nil, err
	}
	return &longChatCs.GetChatDetailResponse{TbLongChat: chat}, nil
}

// 解散群聊
func (c *ChatController) ActionDismissChat(ctx context.Context, req *longChatCs.DismissChatRequest) (*longChatCs.DismissChatResponse, error) {
	uid := middlewares.GetLoginUIDFromContext(ctx)
	if err := longChatBiz.ChatBusiness.DismissChat(req.ChatId, uid); err != nil {
		return nil, err
	}
	return &longChatCs.DismissChatResponse{}, nil
}

// 转让群主
func (c *ChatController) ActionTransferOwner(ctx context.Context, req *longChatCs.TransferOwnerRequest) (*longChatCs.TransferOwnerResponse, error) {
	uid := middlewares.GetLoginUIDFromContext(ctx)
	if err := longChatBiz.ChatBusiness.TransferOwner(req.ChatId, uid, req.NewOwnerUid); err != nil {
		return nil, err
	}
	return &longChatCs.TransferOwnerResponse{}, nil
}
