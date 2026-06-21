package long_chat

import chatModel "github.com/armylong/armylong-go/internal/model/long_chat"

// ==================== 加入聊天 ====================

// 加入聊天-请求
type JoinChatRequest struct {
	ChatId string `json:"chat_id" form:"chat_id"` // 聊天ID
	Uid    int64  `json:"uid" form:"uid"`         // 被拉入的uid
}

// 加入聊天-响应
type JoinChatResponse struct{}

// ==================== 退出聊天 ====================

// 退出聊天-请求
type LeaveChatRequest struct {
	ChatId string `json:"chat_id" form:"chat_id"` // 聊天ID
}

// 退出聊天-响应
type LeaveChatResponse struct{}

// ==================== 踢人 ====================

// 踢人-请求
type KickMemberRequest struct {
	ChatId    string `json:"chat_id" form:"chat_id"`             // 聊天ID
	TargetUid int64  `json:"target_uid" form:"target_uid"`       // 被踢的uid
}

// 踢人-响应
type KickMemberResponse struct{}

// ==================== 成员列表 ====================

// 成员列表-请求
type ListMembersRequest struct {
	ChatId string `json:"chat_id" form:"chat_id"` // 聊天ID
}

// 成员列表-响应
type ListMembersResponse struct {
	List []*chatModel.TbLongChatMember `json:"list"`
}
