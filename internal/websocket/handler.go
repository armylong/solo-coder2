package websocket

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/armylong/armylong-go/internal/common/webcache"
	chatModel "github.com/armylong/armylong-go/internal/model/long_chat"
	"github.com/armylong/armylong-go/internal/middlewares"
	libWs "github.com/armylong/go-library/service/websocket"
	"github.com/gin-gonic/gin"
)

// Manager 全局WebSocket连接管理器
var Manager *libWs.ConnManager

// groupResolver 群组关系解析器，实现 GroupResolver 接口
type groupResolver struct{}

// GetGroups 查询用户参与的所有聊天ID（单聊+群聊），统一走 PushToGroup 频道推送
func (r *groupResolver) GetGroups(uid int64) []string {
	chatIds, err := chatModel.TbLongChatMemberModel.ListChatIdsByUid(uid)
	if err != nil {
		return nil
	}
	return chatIds
}

// Init 初始化WebSocket连接管理器，应在服务启动时调用
func Init() {
	Manager = libWs.NewConnManager(webcache.RedisClient.Client, "armylong:ws", &groupResolver{})
	Manager.Subscribe(context.Background())
}

// HandleWebSocket 处理WebSocket连接请求
// 鉴权方式：通过query参数 ?token=xxx 传入登录token
// 浏览器WebSocket API不支持自定义Header，只能通过URL参数传递
func HandleWebSocket(ctx *gin.Context) {
	token := ctx.Query("token")
	if token == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"message": "缺少token"})
		return
	}

	userInfo, err := middlewares.LoadUserByToken(token)
	if err != nil || userInfo == nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"message": "token无效或已过期"})
		return
	}

	conn, err := libWs.Upgrade(ctx)
	if err != nil {
		return
	}

	// AddConn 自动调用 resolver.GetGroups(uid) 加入群组
	client := Manager.AddConn(userInfo.Uid, conn)
	defer func() {
		// RemoveConn 自动清理群组关系
		Manager.RemoveConn(userInfo.Uid, client)
		_ = client.Close()
	}()

	fmt.Printf("ws connected: uid=%d\n", userInfo.Uid)

	// 读取循环：保持连接，检测断开
	// SetReadDeadline 实现被动心跳检测，60秒无消息则断开
	for {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}

	fmt.Printf("ws disconnected: uid=%d\n", userInfo.Uid)
}
