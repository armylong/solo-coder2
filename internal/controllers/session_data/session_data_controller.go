package session_data

import (
	"context"

	sessionDataBiz "github.com/armylong/armylong-go/internal/business/session_data"
	sessionDataCs "github.com/armylong/armylong-go/internal/cs/session_data"
	"github.com/armylong/armylong-go/internal/middlewares"
)

// 会话数据控制器
type SessionDataController struct{}

// 按key列表拉取会话数据
func (c *SessionDataController) ActionSessionData(ctx context.Context, req *sessionDataCs.SessionDataRequest) (sessionDataCs.SessionDataResponse, error) {
	uid := middlewares.GetLoginUIDFromContext(ctx)
	data, err := sessionDataBiz.SessionDataBusiness.GetSessionData(ctx, uid, req.Keys)
	if err != nil {
		return nil, err
	}
	return data, nil
}
