package long_chat

import (
	"context"

	longChatBiz "github.com/armylong/armylong-go/internal/business/long_chat"
	"github.com/armylong/armylong-go/internal/middlewares"
	longChatCs "github.com/armylong/armylong-go/internal/cs/long_chat"
)

// MemberController 成员管理
type MemberController struct{}

// 加入聊天
func (c *MemberController) ActionJoinChat(ctx context.Context, req *longChatCs.JoinChatRequest) (*longChatCs.JoinChatResponse, error) {
	operatorUid := middlewares.GetLoginUIDFromContext(ctx)
	chat, err := longChatBiz.ChatBusiness.GetChat(req.ChatId)
	if err != nil || chat == nil {
		return nil, err
	}
	// 只有群主才能拉人
	if chat.OwnerUid != operatorUid {
		return nil, err
	}
	if err := longChatBiz.MemberBusiness.JoinChat(req.ChatId, req.Uid); err != nil {
		return nil, err
	}
	return &longChatCs.JoinChatResponse{}, nil
}

// 退出聊天
func (c *MemberController) ActionLeaveChat(ctx context.Context, req *longChatCs.LeaveChatRequest) (*longChatCs.LeaveChatResponse, error) {
	uid := middlewares.GetLoginUIDFromContext(ctx)
	if err := longChatBiz.MemberBusiness.LeaveChat(req.ChatId, uid); err != nil {
		return nil, err
	}
	return &longChatCs.LeaveChatResponse{}, nil
}

// 踢人
func (c *MemberController) ActionKickMember(ctx context.Context, req *longChatCs.KickMemberRequest) (*longChatCs.KickMemberResponse, error) {
	uid := middlewares.GetLoginUIDFromContext(ctx)
	if err := longChatBiz.MemberBusiness.KickMember(req.ChatId, uid, req.TargetUid); err != nil {
		return nil, err
	}
	return &longChatCs.KickMemberResponse{}, nil
}

// 成员列表
func (c *MemberController) ActionListMembers(ctx context.Context, req *longChatCs.ListMembersRequest) (*longChatCs.ListMembersResponse, error) {
	list, err := longChatBiz.MemberBusiness.ListMembers(req.ChatId)
	if err != nil {
		return nil, err
	}
	return &longChatCs.ListMembersResponse{List: list}, nil
}
