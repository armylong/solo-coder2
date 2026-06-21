# armylong-go

Go 语言 Web 应用，基于 gin + sqlite + cli 构建。

## 技术栈

| 组件 | 技术 |
|---|---|
| Web 框架 | gin |
| 数据库 | sqlite |
| 命令行 | urfave/cli/v2 |
| 认证 | JWT (golang-jwt/jwt/v5) |

## 启动

```bash
go run main.go
```

默认启动 Web 服务。

## 命令行工具

```bash
# 查看所有命令
go run main.go help

# 创建超级管理员（冷启动时使用）
go run main.go create-super <账号> <密码>

# 示例
go run main.go create-super armylong mypassword123
```

### 冷启动说明

项目首次运行时数据库为空，没有任何用户和权限数据。需要通过命令行创建第一个超级管理员：

```bash
# 1. 启动一次项目，初始化数据库
go run main.go

# 2. Ctrl+C 停掉，创建超管
go run main.go create-super armylong mypassword123

# 3. 重新启动，用超管账号登录
go run main.go
```

如果账号已存在，会自动提升为超级管理员，不会重复创建。
