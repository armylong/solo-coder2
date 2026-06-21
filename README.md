<p align="center">
  <h1 align="center">Armvault</h1>
  <p align="center">全栈 Web 应用 · 命令行工具集 · Docker 容器编排</p>
  <p align="center">
    <img src="https://img.shields.io/badge/Go-1.25-00ADD8?style=flat-square&logo=go" alt="Go" />
    <img src="https://img.shields.io/badge/Gin-Web_Framework-008EC4?style=flat-square" alt="Gin" />
    <img src="https://img.shields.io/badge/Docker-Containerized-2496ED?style=flat-square&logo=docker" alt="Docker" />
    <img src="https://img.shields.io/badge/Feishu-API-3370FF?style=flat-square" alt="Feishu" />
    <img src="https://img.shields.io/badge/SQLite-Database-003B57?style=flat-square&logo=sqlite" alt="SQLite" />
  </p>
</p>

---

## Overview

Armvault 是一个基于 Go 的全栈 Web 应用平台，集成了 Web 服务、CLI 工具链和 Docker 容器管理能力。采用前后端一体架构，后端 Go + Gin 提供 API，前端原生 HTML/CSS/JS 渲染交互界面，通过 SQLite 持久化数据，飞书开放平台实现数据协同。

## Architecture

```
┌─────────────────────────────────────────────────┐
│                    main.go                       │
│              serve / cli 双模式入口               │
├──────────────────────┬──────────────────────────┤
│       Web 层         │        CLI 层             │
│   Gin + Longgin      │     urfave/cli/v2         │
│   路由 / 中间件       │     命令 / 子命令          │
├──────────────────────┼──────────────────────────┤
│    Controller        │        Cmd                │
│   请求分发与参数校验    │   独立命令行任务           │
├──────────────────────┴──────────────────────────┤
│                  Business                        │
│             业务逻辑层 (核心)                      │
├─────────────────────────────────────────────────┤
│                    CS                            │
│            外部服务调用 (飞书/高德/Redis)           │
├─────────────────────────────────────────────────┤
│                   Model                          │
│           数据模型 (SQLite ORM)                   │
├─────────────────────────────────────────────────┤
│              Common / Config                     │
│            公共组件与配置                          │
└─────────────────────────────────────────────────┘
```

## Features

### Web 应用

| 模块 | 说明 | 权限 |
|------|------|------|
| 桌面系统 | 可定制化桌面布局，应用图标拖拽排列 | 登录用户 |
| 拼拼坐 | 拼车出行平台，司机/乘客/运营管理 | 登录用户 / 管理员 |
| 氧分 | 虚拟积分系统，充值/消费/转账/过期清理 | 登录用户 |
| 长文档 | 基于 LLM 的长文档生成与管理 | 登录用户 |
| 长仓库 | 文件存储与分享 | 登录用户 |
| 高德地图 | 地理编码 / 逆地理编码 / 路线规划代理 | 登录用户 |
| API 抓包 | HTTP 请求抓取与回放 | 公开 |
| 系统监控 | CPU / 内存 / 磁盘 / 网络 / 进程实时监控 | 超级管理员 |
| 用户管理 | 注册 / 登录 / JWT 认证 / 权限分级 | 分级控制 |

### CLI 工具

| 命令 | 说明 |
|------|------|
| `serve` | 启动 Web 服务 |
| `super <account> <password>` | 创建超级管理员 |
| `coding700 init/reset/collect/upload` | Docker 容器化编程环境管理 |
| `solo session/upload` | Solo Coder 会话管理与飞书同步 |
| `monitor` | 终端实时系统监控面板 |
| `rubrics download/upload` | 评分数据下载与上传 |
| `yangfen` | 氧分管理（充值/消费/转账/查询） |
| `todo` | 任务管理（增删改查） |
| `obfuscate --path <dir>` | Go 代码混淆（防查重） |
| `redis get/set` | Redis 快捷操作 |
| `demo` | 演示模块 |

### 代码混淆

内置 Go 代码混淆工具，支持以下变换：

- 标识符重命名（未导出标识符，同长度随机替换）
- 注释全量删除
- 字符串字面量拆分
- 表达式变换（`true` → `!false`，比较运算符交换）
- 废代码注入（随机方法 + 包级变量）
- 声明顺序重排
- 编译验证 + 自动回滚

## Tech Stack

| 层级 | 技术 |
|------|------|
| 语言 | Go 1.25 |
| Web 框架 | Gin + Longgin (自研封装) |
| CLI 框架 | urfave/cli/v2 |
| 数据库 | SQLite (go-library ORM) |
| 缓存 | Redis |
| 认证 | JWT (golang-jwt) |
| 外部 API | 飞书开放平台 / 高德地图 |
| 系统监控 | gopsutil |
| 容器化 | Docker |
| 前端 | 原生 HTML / CSS / JavaScript |

## Project Structure

```
armylong-go/
├── main.go                    # 入口
├── internal/
│   ├── business/              # 业务逻辑层
│   │   ├── index/             #   首页/桌面
│   │   ├── ppz/               #   拼拼坐
│   │   ├── yangfen/           #   氧分
│   │   ├── long_doc/          #   长文档
│   │   ├── long_store/        #   长仓库
│   │   ├── work/              #   工作管理
│   │   ├── monitor/           #   系统监控
│   │   ├── feishu/            #   飞书服务
│   │   ├── gaode/             #   高德地图
│   │   ├── user/              #   用户
│   │   └── ...
│   ├── controllers/           # 控制器层
│   ├── cs/                    # 外部服务调用层
│   ├── model/                 # 数据模型层
│   ├── cmd/                   # CLI 命令
│   ├── common/                # 公共组件 (config/webcache)
│   ├── middlewares/           # 中间件 (auth)
│   ├── register_cmd.go        # CLI 命令注册
│   └── register_web.go        # Web 路由注册
├── static/                    # 前端静态资源
│   ├── _common/               #   公共组件 (登录/侧边栏)
│   ├── _index/                #   桌面框架
│   ├── ppz/                   #   拼拼坐前端
│   ├── yangfen/               #   氧分前端
│   ├── long-doc/              #   长文档前端
│   ├── chinese-chess/         #   中国象棋
│   ├── doudizhu/              #   斗地主
│   ├── gomoku/                #   五子棋
│   └── ...
├── docs/                      # 项目文档
├── Dockerfile                 # Docker 构建
├── go.mod
└── go.sum
```

## Quick Start

### 环境要求

- Go 1.25+
- Redis (可选)
- Docker (容器管理功能需要)

### 安装运行

```bash
# 克隆项目
git clone https://github.com/armylong/armylong-go.git
cd armylong-go

# 安装依赖
go mod tidy

# 启动 Web 服务
go run main.go serve

# 或构建后运行
go build -o app .
./app serve
```

### 创建管理员

```bash
./app super admin 123456
```

### Docker 部署

```bash
docker build -t armvault .
docker run -p 8080:80 armvault
```

## Statistics

| 类型 | 文件数 | 代码行数 |
|------|--------|----------|
| Go | 126 | 19,368 |
| JavaScript | 95 | 24,832 |
| CSS | 40 | 12,350 |
| HTML | 35 | 3,235 |
| **合计** | **296** | **59,785** |

## License

Private Project
