package internal // 包名=目录名，适配internal/根目录

import (
	"github.com/armylong/armylong-go/internal/cmd"
	"github.com/armylong/go-library/service/command"
	"github.com/urfave/cli/v2"
)

// RegisterCmd 集中注册所有子命令（修正所有Cli语法错误）
func RegisterCmd(command command.BaseCommand) {

	command.AddCliCommand(&cli.Command{
		Name:      "super",
		Usage:     "创建超级管理员",
		Action:    cmd.CreateSuperHandler,
		ArgsUsage: "<账号> <密码>",
	})

	command.AddCliCommand(&cli.Command{
		Name:   `demo`,
		Usage:  `演示参数接收`,
		Action: cmd.DemoHandler,
		Subcommands: []*cli.Command{
			{
				Name:   "hello",
				Usage:  "hello",
				Action: cmd.HelloHandler,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "id",
						Usage: "ID",
					},
				},
			},
			{
				Name:   "redis",
				Usage:  "redis",
				Action: cmd.RedisHandler,
			},
			{
				Name:   "feishu",
				Usage:  "飞书",
				Action: cmd.FeishuHandler,
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "sex",
				Usage: "性别",
			},
		},
	})

	command.AddCliCommand(&cli.Command{
		Name:   "redis",
		Usage:  "redis",
		Action: cmd.RedisHandler,
		Subcommands: []*cli.Command{
			{
				Name:      "set",
				Usage:     "设置",
				Action:    cmd.RedisSetHandler,
				ArgsUsage: "<key> <value> <expire>",
			},
			{
				Name:      "get",
				Usage:     "获取",
				Action:    cmd.RedisGetHandler,
				ArgsUsage: "<key>",
			},
		},
	})

	command.AddCliCommand(&cli.Command{
		Name:   "todo",
		Action: cmd.TodoHandler,
		Usage:  "任务管理",
		Flags: []cli.Flag{
			&cli.Int64Flag{
				Name:  "task_id",
				Usage: "任务ID（可选）",
			},
			&cli.StringFlag{
				Name:  "title",
				Usage: "任务标题（create时必填）",
			},
			&cli.StringFlag{
				Name:  "desc",
				Usage: "任务描述（create时必填）",
			},
			&cli.Int64Flag{
				Name:  "sort",
				Usage: "任务排序值，数字越大越靠前（可选）",
			},
			&cli.StringFlag{
				Name:  "expire_at",
				Usage: "过期时间，格式：2006-01-02 15:04:05（可选）",
			},
		},
	})

	command.AddCliCommand(&cli.Command{
		Name:   "yangfen",
		Usage:  "氧分管理",
		Action: cmd.YangfenCmd.YangfenHandler,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "uid",
				Usage: "用户ID",
			},
			&cli.IntFlag{
				Name:  "amount",
				Usage: "金额",
			},
			&cli.StringFlag{
				Name:  "to-uid",
				Usage: "转账目标用户ID",
			},
			&cli.Int64Flag{
				Name:  "expire-sec",
				Usage: "过期时间（秒）",
			},
			&cli.StringFlag{
				Name:  "transaction-id",
				Usage: "交易ID",
			},
		},
	})

	command.AddCliCommand(&cli.Command{
		Name:   "monitor",
		Usage:  "系统监控",
		Action: cmd.MonitorCmd.MonitorHandler,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "refresh",
				Usage: "实时刷新显示",
			},
			&cli.IntFlag{
				Name:  "interval",
				Usage: "刷新间隔（秒）",
				Value: 2,
			},
			&cli.StringFlag{
				Name:  "sort",
				Usage: "排序方式（cpu/memory/pid）",
			},
			&cli.IntFlag{
				Name:  "limit",
				Usage: "显示数量限制",
				Value: 10,
			},
		},
	})

	command.AddCliCommand(&cli.Command{
		Name:   "obfuscate",
		Usage:  "混淆Go代码",
		Action: cmd.GoObfuscateHandler,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "path",
				Usage: "项目路径",
			},
		},
	})

}
