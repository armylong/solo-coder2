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
		Name:   "doubao_pro_testing",
		Usage:  "豆包专业版测试",
		Action: cmd.DoubaoProTestingCmd.DoubaoProTestingHandler,
	})

	command.AddCliCommand(&cli.Command{
		Name:   "refresh_works",
		Usage:  "刷新工作",
		Action: cmd.RefreshWorksHandler,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "work_space",
				Usage: "工作空间",
			},
			&cli.StringFlag{
				Name:  "has_file_names",
				Usage: "包含的文件",
			},
			&cli.StringFlag{
				Name:  "no_has_file_names",
				Usage: "不包含的文件",
			},
		},
	})

	command.AddCliCommand(&cli.Command{
		Name:   "bad_pattern_testing",
		Usage:  "坏模式测试",
		Action: cmd.BadPatternTestingCmd.BadPatternTestingHandler,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "question_id",
				Usage: "题目ID",
			},
			&cli.StringFlag{
				Name:  "work_path",
				Usage: "工作目录",
			},
		},
	})

	command.AddCliCommand(&cli.Command{
		Name:   "dogfooding_testing",
		Usage:  "dogfooding测试",
		Action: cmd.DogfoodingTestingCmd.DogfoodingTestingHandler,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "question_id",
				Usage: "题目ID",
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
		Name:  "solo",
		Usage: "solo coder",
		Subcommands: []*cli.Command{
			{
				Name:   "session",
				Usage:  "solo coder session",
				Action: cmd.SoloCoderSessionHandler,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "id",
						Usage: "题目ID",
					},
				},
			},
			{
				Name:   "upload",
				Usage:  "上传solo coder 上传至飞书表格",
				Action: cmd.SoloCoderUploadFeishuHandler,
			},
		},
	})

	command.AddCliCommand(&cli.Command{
		Name:   `rubrics`,
		Usage:  `前端rubrics评分`,
		Action: cmd.RubricsHandler,
		Subcommands: []*cli.Command{
			{
				Name:   "download",
				Usage:  "下载rubrics评分",
				Action: cmd.RubricsDownloadWhileHandler,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "work_home",
						Usage: "工作目录",
					},
				},
			},
			{
				Name:   "upload",
				Usage:  "上传rubrics评分",
				Action: cmd.RubricsUploadWhileHandler,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "work_home",
						Usage: "工作目录",
					},
				},
			},
		},
	})

	command.AddCliCommand(&cli.Command{
		Name:   `coding700`,
		Usage:  `coding700 评分`,
		Action: cmd.Coding700Handler,
		Subcommands: []*cli.Command{
			{
				Name:   "new",
				Usage:  "初始化",
				Action: cmd.Coding700NewHandler,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "repo_url",
						Usage: "仓库URL",
					},
				},
			},
			{
				Name:   "init",
				Usage:  "初始化",
				Action: cmd.Coding700InitHandler,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "project",
						Usage: "项目名称",
					},
					&cli.BoolFlag{
						Name:  "git_init",
						Usage: "初始化git仓库",
						Value: false,
					},
				},
			},
			{
				Name:   "prompt_init",
				Usage:  "初始化题目环境",
				Action: cmd.Coding700ResetHandler,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "project", // 处理rollout_id的时候才有用
						Usage: "项目名称",
					},
					&cli.StringFlag{
						Name:  "prompt_id",
						Usage: "题目ID列表, 逗号分割",
					},
					&cli.IntFlag{
						Name:  "rollout_id", // 重置题目环境前是否等待指定的rollout.json文件处理完成
						Usage: "子任务ID",
					},
				},
			},
			{
				Name:   "collect",
				Usage:  "收集任务成果",
				Action: cmd.Coding700CollectTaskResultsHandler,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "project",
						Usage: "项目名称",
					},
					&cli.IntFlag{
						Name:  "prompt_id",
						Usage: "题目ID",
					},
					&cli.IntFlag{
						Name:  "rollout_id",
						Usage: "子任务ID",
					},
				},
			},
			{
				Name:   "upload",
				Usage:  "上传任务成果",
				Action: cmd.Coding700UploadTaskDataHandler,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "project",
						Usage: "项目名称",
					},
					&cli.BoolFlag{
						Name:  "prompt_yn",
						Usage: "是否检查prompt.json文件",
						Value: true,
					},
					&cli.BoolFlag{
						Name:  "rollout_yn",
						Usage: "是否检查rollout.json文件",
						Value: true,
					},
					&cli.StringFlag{
						Name:  "repo_ignore_fields",
						Usage: "要忽略的字段",
					},
					&cli.BoolFlag{
						Name:  "clean_install_modules",
						Usage: "是否清理安装模块",
						Value: true,
					},
				},
			},
			{
				Name:   "upload_init_repo",
				Usage:  "上传初始化仓库到飞书表格",
				Action: cmd.Coding700UploadInitRepoHandler,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "repo_type",
						Usage: "仓库类型",
					},
					&cli.StringFlag{
						Name:  "repo_url",
						Usage: "仓库URL",
					},
				},
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
