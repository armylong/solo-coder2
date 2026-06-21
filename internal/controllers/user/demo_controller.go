package user

import (
	demoBiz "github.com/armylong/armylong-go/internal/business/demo"
	userCs "github.com/armylong/armylong-go/internal/cs/user"
	"github.com/gin-gonic/gin"
)

// DemoController 示例接口
type DemoController struct{}

// 获取示例消息
func (c *DemoController) ActionHello(ctx *gin.Context) (res *userCs.DemoMessage, err error) {
	message, err := demoBiz.DemoBusiness.GetMessage(ctx)
	if err != nil || message == "" {
		return res, err
	}

	return &userCs.DemoMessage{
		Message: message,
	}, nil
}

// 设置示例消息
func (c *DemoController) ActionSetHello(ctx *gin.Context, request *userCs.DemoMessage) (res *userCs.DemoMessage, err error) {
	message, err := demoBiz.DemoBusiness.SetMessage(ctx, request.Message)
	if err != nil || message == "" {
		return res, err
	}

	return &userCs.DemoMessage{
		Message: message,
	}, nil
}
