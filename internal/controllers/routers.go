package controllers

import (
	"errors"
	"net/http"

	apiCatcherController "github.com/armylong/armylong-go/internal/controllers/api_catcher"
	"github.com/armylong/armylong-go/internal/controllers/index"
	longChatController "github.com/armylong/armylong-go/internal/controllers/long_chat"
	longDocController "github.com/armylong/armylong-go/internal/controllers/long_doc"
	monitorController "github.com/armylong/armylong-go/internal/controllers/monitor"
	sessionDataController "github.com/armylong/armylong-go/internal/controllers/session_data"
	"github.com/armylong/armylong-go/internal/controllers/settings"
	"github.com/armylong/armylong-go/internal/controllers/sqlite_long"
	userController "github.com/armylong/armylong-go/internal/controllers/user"
	"github.com/armylong/armylong-go/internal/controllers/yangfen"
	"github.com/armylong/armylong-go/internal/middlewares"
	ws "github.com/armylong/armylong-go/internal/websocket"
	"github.com/armylong/go-library/service/longgin"
	"github.com/gin-gonic/gin"
)

// 注册所有路由
func RegisterRouters(engine *gin.Engine) {

	// ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓ WebSocket ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓
	ws.Init()
	engine.GET("/ws", ws.HandleWebSocket)
	// ↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑ WebSocket ↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑

	// ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓ 后端接口 ↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓↓
	engine.Any(`/`, homepage)

	engine.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(http.StatusNotFound, longgin.ErrorWithContext(ctx, errors.New("not found"), longgin.CodeNotFound))
	})

	// ==================== 无需登录 ====================
	authGroup := engine.Group("/auth")
	longgin.RegisterJsonController(authGroup, &userController.AuthController{})

	apiCatcherRoot := engine.Group("/api_catcher")
	longgin.RegisterJsonController(apiCatcherRoot, &apiCatcherController.ApiCatcherController{})

	// ==================== 登录即可访问 ====================
	publicGroup := engine.Group("", middlewares.Middleware)

	longgin.RegisterJsonController(publicGroup.Group("/session_data"), &sessionDataController.SessionDataController{})
	longgin.RegisterJsonController(publicGroup.Group("/user"), &userController.UserController{})
	longgin.RegisterJsonController(publicGroup.Group("/user/demo"), &userController.DemoController{})
	longgin.RegisterJsonController(publicGroup.Group("/yangfen"), &yangfen.YangfenController{})
	longgin.RegisterJsonController(publicGroup.Group("/index"), &index.IndexController{})
	longgin.RegisterJsonController(publicGroup.Group("/long_doc"), &longDocController.LongDocController{})
	longgin.RegisterJsonController(publicGroup.Group("/long_chat/chat"), &longChatController.ChatController{})
	longgin.RegisterJsonController(publicGroup.Group("/long_chat/message"), &longChatController.MessageController{})
	longgin.RegisterJsonController(publicGroup.Group("/long_chat/friend"), &longChatController.FriendController{})
	longgin.RegisterJsonController(publicGroup.Group("/long_chat/member"), &longChatController.MemberController{})

	// ==================== 管理员及以上 ====================
	adminGroup := engine.Group("", middlewares.Middleware, middlewares.RequireAdmin)

	longgin.RegisterJsonController(adminGroup.Group("/user_management"), &userController.UserManagementController{})

	// ==================== 仅超级管理员 ====================
	superAdminGroup := engine.Group("", middlewares.Middleware, middlewares.RequireSuperAdmin)

	longgin.RegisterJsonController(superAdminGroup.Group("/monitor"), &monitorController.MonitorController{})
	longgin.RegisterJsonController(superAdminGroup.Group("/sqlite_long"), &sqlite_long.SqliteLongController{})
	longgin.RegisterJsonController(superAdminGroup.Group("/settings"), &settings.SettingsController{})
}

// ↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑ 后端接口 ↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑↑
