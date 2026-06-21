package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 首页
func homepage(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "index.gohtml", map[string]any{
		"user": nil,
	})
}
