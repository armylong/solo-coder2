package long_chat

import (
	"context"

	longChatBiz "github.com/armylong/armylong-go/internal/business/long_chat"
	"github.com/armylong/armylong-go/internal/middlewares"
	longChatCs "github.com/armylong/armylong-go/internal/cs/long_chat"
	userModel "github.com/armylong/armylong-go/internal/model/user"
)

// FriendController 好友管理
type FriendController struct{}

// 添加好友
func (c *FriendController) ActionAddFriend(ctx context.Context, req *longChatCs.AddFriendRequest) (*longChatCs.AddFriendResponse, error) {
	uid := middlewares.GetLoginUIDFromContext(ctx)
	if err := longChatBiz.FriendBusiness.AddFriend(uid, req.FriendUid); err != nil {
		return nil, err
	}
	return &longChatCs.AddFriendResponse{}, nil
}

// 同意好友
func (c *FriendController) ActionAcceptFriend(ctx context.Context, req *longChatCs.AcceptFriendRequest) (*longChatCs.AcceptFriendResponse, error) {
	uid := middlewares.GetLoginUIDFromContext(ctx)
	if err := longChatBiz.FriendBusiness.AcceptFriend(uid, req.FriendUid); err != nil {
		return nil, err
	}
	return &longChatCs.AcceptFriendResponse{}, nil
}

// 拒绝好友
func (c *FriendController) ActionRejectFriend(ctx context.Context, req *longChatCs.RejectFriendRequest) (*longChatCs.RejectFriendResponse, error) {
	uid := middlewares.GetLoginUIDFromContext(ctx)
	if err := longChatBiz.FriendBusiness.RejectFriend(uid, req.FriendUid); err != nil {
		return nil, err
	}
	return &longChatCs.RejectFriendResponse{}, nil
}

// 删除好友
func (c *FriendController) ActionDeleteFriend(ctx context.Context, req *longChatCs.DeleteFriendRequest) (*longChatCs.DeleteFriendResponse, error) {
	uid := middlewares.GetLoginUIDFromContext(ctx)
	if err := longChatBiz.FriendBusiness.DeleteFriend(uid, req.FriendUid); err != nil {
		return nil, err
	}
	return &longChatCs.DeleteFriendResponse{}, nil
}

// 好友列表
func (c *FriendController) ActionListFriends(ctx context.Context, req *longChatCs.ListFriendsRequest) (*longChatCs.ListFriendsResponse, error) {
	uid := middlewares.GetLoginUIDFromContext(ctx)
	friends, err := longChatBiz.FriendBusiness.ListFriends(uid)
	if err != nil {
		return nil, err
	}

	list := make([]*longChatCs.FriendItem, 0, len(friends))
	for _, f := range friends {
		name := ""
		u, err := userModel.TbUserModel.GetByUid(f.FriendUid)
		if err == nil && u != nil {
			name = u.Name
		}
		list = append(list, &longChatCs.FriendItem{
			Uid:       f.Uid,
			FriendUid: f.FriendUid,
			Name:      name,
			Status:    f.Status,
			CreatedAt: f.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &longChatCs.ListFriendsResponse{List: list}, nil
}

// 待确认好友请求
func (c *FriendController) ActionListPendingRequests(ctx context.Context, req *longChatCs.ListPendingRequestsRequest) (*longChatCs.ListPendingRequestsResponse, error) {
	uid := middlewares.GetLoginUIDFromContext(ctx)
	pendings, err := longChatBiz.FriendBusiness.ListPendingRequests(uid)
	if err != nil {
		return nil, err
	}

	list := make([]*longChatCs.PendingRequestItem, 0, len(pendings))
	for _, p := range pendings {
		name := ""
		u, err := userModel.TbUserModel.GetByUid(p.Uid)
		if err == nil && u != nil {
			name = u.Name
		}
		list = append(list, &longChatCs.PendingRequestItem{
			Uid:       p.Uid,
			FriendUid: p.FriendUid,
			Name:      name,
			Status:    p.Status,
			CreatedAt: p.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return &longChatCs.ListPendingRequestsResponse{List: list}, nil
}

// 搜索用户
func (c *FriendController) ActionSearch(ctx context.Context, req *longChatCs.SearchRequest) (*longChatCs.SearchResponse, error) {
	uid := middlewares.GetLoginUIDFromContext(ctx)
	items, err := longChatBiz.FriendBusiness.SearchUser(uid, req.Keyword)
	if err != nil {
		return nil, err
	}

	list := make([]*longChatCs.SearchItem, 0, len(items))
	for _, item := range items {
		list = append(list, &longChatCs.SearchItem{
			Uid:      item.Uid,
			Account:  item.Account,
			Name:     item.Name,
			IsFriend: item.IsFriend,
		})
	}
	return &longChatCs.SearchResponse{List: list}, nil
}
