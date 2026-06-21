package long_chat

import chatModel "github.com/armylong/armylong-go/internal/model/long_chat"

// ==================== 发送消息 ====================

// 发送消息-请求
type SendMessageRequest struct {
	ChatId  string `json:"chat_id" form:"chat_id"`     // 聊天ID
	MsgType string `json:"msg_type" form:"msg_type"`   // 消息类型: text/image/voice
	Content string `json:"content" form:"content"`     // 消息内容
}

// 发送消息-响应
type SendMessageResponse struct {
	Msg *chatModel.TbLongChatMessage `json:"msg"`
}

// ==================== 历史消息 ====================

// 历史消息-请求
type ListMessagesRequest struct {
	ChatId string `json:"chat_id" form:"chat_id"` // 聊天ID
	Limit  int    `json:"limit" form:"limit"`     // 每页条数，默认20
	Offset int    `json:"offset" form:"offset"`   // 偏移量
}

// 消息项（带发送者昵称）
type MessageItem struct {
	MsgId      string `json:"msg_id"`
	ChatId     string `json:"chat_id"`
	FromUid    int64  `json:"from_uid"`
	SenderName string `json:"sender_name"` // 发送者昵称
	MsgType    string `json:"msg_type"`
	Content    string `json:"content"`
	CreatedAt  string `json:"created_at"`
}

// 历史消息-响应
type ListMessagesResponse struct {
	Uid  int64          `json:"uid"`  // 当前用户uid，前端用于判断消息方向
	List []*MessageItem `json:"list"`
}

// ==================== 清零未读 ====================

// 清零未读-请求
type ClearUnreadRequest struct {
	ChatId string `json:"chat_id" form:"chat_id"` // 聊天ID
}

// 清零未读-响应
type ClearUnreadResponse struct{}
