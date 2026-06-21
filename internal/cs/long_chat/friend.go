package long_chat

// ==================== 添加好友 ====================

// 添加好友-请求
type AddFriendRequest struct {
	FriendUid int64 `json:"friend_uid" form:"friend_uid"` // 对方uid
}

// 添加好友-响应
type AddFriendResponse struct{}

// ==================== 同意好友 ====================

// 同意好友-请求
type AcceptFriendRequest struct {
	FriendUid int64 `json:"friend_uid" form:"friend_uid"` // 对方uid
}

// 同意好友-响应
type AcceptFriendResponse struct{}

// ==================== 拒绝好友 ====================

// 拒绝好友-请求
type RejectFriendRequest struct {
	FriendUid int64 `json:"friend_uid" form:"friend_uid"` // 对方uid
}

// 拒绝好友-响应
type RejectFriendResponse struct{}

// ==================== 删除好友 ====================

// 删除好友-请求
type DeleteFriendRequest struct {
	FriendUid int64 `json:"friend_uid" form:"friend_uid"` // 对方uid
}

// 删除好友-响应
type DeleteFriendResponse struct{}

// ==================== 好友列表 ====================

// 好友列表-请求
type ListFriendsRequest struct{}

// 好友列表-响应
type ListFriendsResponse struct {
	List []*FriendItem `json:"list"`
}

// 好友列表项
type FriendItem struct {
	Uid       int64  `json:"uid"`        // 发起方uid
	FriendUid int64  `json:"friend_uid"` // 被添加方uid
	Name      string `json:"name"`       // 好友昵称
	Status    int    `json:"status"`     // 状态
	CreatedAt string `json:"created_at"`
}

// ==================== 待确认好友请求 ====================

// 待确认好友请求-请求
type ListPendingRequestsRequest struct{}

// 待确认好友请求-响应
type ListPendingRequestsResponse struct {
	List []*PendingRequestItem `json:"list"`
}

// 待确认好友请求项
type PendingRequestItem struct {
	Uid       int64  `json:"uid"`        // 发起方uid
	FriendUid int64  `json:"friend_uid"` // 被添加方uid
	Name      string `json:"name"`       // 请求者昵称
	Status    int    `json:"status"`     // 状态
	CreatedAt string `json:"created_at"`
}

// ==================== 搜索用户 ====================

// 搜索用户-请求
type SearchRequest struct {
	Keyword string `json:"keyword" form:"keyword"` // 搜索关键词
}

// 搜索用户结果项
type SearchItem struct {
	Uid      int64  `json:"uid"`       // 用户uid
	Account  string `json:"account"`   // 账号
	Name     string `json:"name"`      // 昵称
	IsFriend bool   `json:"is_friend"` // 是否已是好友
}

// 搜索用户-响应
type SearchResponse struct {
	List []*SearchItem `json:"list"`
}
