package long_chat

import chatModel "github.com/armylong/armylong-go/internal/model/long_chat"

// ==================== 聊天列表 ====================

// 聊天列表-请求
type GetChatListRequest struct{}

// 聊天列表-响应
type GetChatListResponse struct {
	List []*ChatListItem `json:"list"`
}

// 聊天列表项
type ChatListItem struct {
	ChatId    string `json:"chat_id"`
	ChatType  int    `json:"chat_type"`
	ChatName  string `json:"chat_name"`
	Unread    int    `json:"unread"`
	LastMsg   string `json:"last_msg"`
	LastMsgAt string `json:"last_msg_at"`
}

// ==================== 创建单聊 ====================

// 创建单聊-请求
type CreatePrivateChatRequest struct {
	TargetUid int64 `json:"target_uid" form:"target_uid"` // 对方uid
}

// 创建单聊-响应
type CreatePrivateChatResponse struct {
	ChatId string `json:"chat_id"`
}

// ==================== 创建群聊 ====================

// 创建群聊-请求
type CreateGroupRequest struct {
	GroupName  string  `json:"group_name" form:"group_name"`     // 群名
	MemberUids []int64 `json:"member_uids" form:"member_uids"`   // 成员uid列表
}

// 创建群聊-响应
type CreateGroupResponse struct {
	ChatId string `json:"chat_id"`
}

// ==================== 获取聊天详情 ====================

// 聊天详情-请求
type GetChatDetailRequest struct {
	ChatId string `json:"chat_id" form:"chat_id"`
}

// 聊天详情-响应
type GetChatDetailResponse struct {
	*chatModel.TbLongChat
}

// ==================== 解散群聊 ====================

// 解散群聊-请求
type DismissChatRequest struct {
	ChatId string `json:"chat_id" form:"chat_id"`
}

// 解散群聊-响应
type DismissChatResponse struct{}

// ==================== 转让群主 ====================

// 转让群主-请求
type TransferOwnerRequest struct {
	ChatId      string `json:"chat_id" form:"chat_id"`
	NewOwnerUid int64  `json:"new_owner_uid" form:"new_owner_uid"` // 新群主uid
}

// 转让群主-响应
type TransferOwnerResponse struct{}
