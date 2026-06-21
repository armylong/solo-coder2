package internal

import (
	"github.com/armylong/armylong-go/internal/controllers"
	"github.com/armylong/go-library/service/command"
	"github.com/armylong/go-library/service/longgin"
	"github.com/gin-gonic/gin"

	"github.com/urfave/cli/v2"
)

func RegisterWeb(command command.BaseCommand) {
	command.AddCliCommand(&cli.Command{
		Name:    "serve",
		Aliases: []string{"web"},
		Action: func(ctx *cli.Context) error {
			return longgin.Start(func(engine *gin.Engine) {
				//engine.Use(middlewares.LogPostData()) // 在日志中记录post参数
				controllers.RegisterRouters(engine)
			})
		},
	})
}
