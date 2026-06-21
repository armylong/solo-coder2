package session_data

// 会话数据请求
type SessionDataRequest struct {
	Keys []string `json:"keys"` // 要获取的数据key列表
}

// 会话数据响应
type SessionDataResponse = map[string]any
