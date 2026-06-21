# WebSocket 接入调研与实现

## 背景

当前匹配中订单、行程状态等场景使用前端轮询（5s 间隔请求接口），体验差且浪费资源。需要引入 WebSocket 实现服务端主动推送。

## 技术选型

- `github.com/gorilla/websocket` — Go 生态最成熟的 WS 库
- `github.com/redis/go-redis/v9` — 已有 Redis 基础设施，用于 Pub/Sub 跨实例推送

## 架构

```
客户端 ←── WebSocket ──→ Gin (/ws)
                              │
                        WS Handler (升级 + JWT 鉴权)
                              │
                        ConnManager (内存 map + Redis Pub/Sub)
                        uid → []*ClientConn
                        groupId → 在线成员uid集合
                              │
                   PushToUsers(uids, msg)       PushToGroup(groupId, msg)
                   单播/小群，走全局频道          大群，走群频道隔离
                        │                            │
                        ├── pushLocal                ├── pushLocalGroup（查本地 groups map）
                        │                            │
                        └── Redis PUBLISH 全局频道    └── Redis PUBLISH 群频道
                              → 其他实例遍历 uids          → 订阅了该群的实例收到
                              → pushLocal                  → 查本地 groups map pushLocal
```

- **单播/小群**：`PushToUsers(uids, msg)`，走全局频道，消息体携带 uid 列表
- **大群**：`PushToGroup(groupId, msg)`，走群频道隔离，消息体只带 groupId，各实例查本地 groups map 推送
- **单实例**：本地直接推送，Redis Pub/Sub 仅多一步 publish，开销极小
- **多实例**：通过 Redis Pub/Sub 广播，其他实例收到消息后推送给本地连接
- **无 Redis**：redisClient 传 nil，退化为纯内存模式
- **无 GroupResolver**：传 nil，退化为纯单播模式

## 推送模式

### PushToUsers — 单播/小群

```go
// 单播：传一个 uid
manager.PushToUsers([]int64{123}, msg)

// 小群播：传多个 uid
manager.PushToUsers([]int64{1, 2, 3}, msg)
```

走全局频道 `armylong:ws`，Redis 消息体携带 uid 列表。

### PushToGroup — 大群

```go
// 群播：传 groupId
manager.PushToGroup("G1", msg)
```

走群频道 `armylong:ws:group:G1`，Redis 消息体只带 groupId，各实例查本地 groups map 推送。

**为什么大群不能用 PushToUsers？** 万人群的 uid 列表序列化后几十 KB，每条消息都带，Redis 和网络压力大。频道隔离后消息体只有 groupId 几十字节，各实例查本地内存 map 即可。

## 群组管理

### GroupResolver 回调接口

业务层实现 `GetGroups(uid) []string`，WebSocket 层自动管理群组生命周期：

```go
type GroupResolver interface {
    GetGroups(uid int64) []string
}
```

- 用户连接时：`AddConn` 自动调用 `resolver.GetGroups(uid)`，将用户加入对应群组
- 用户断开时：`RemoveConn` 自动将用户从所有群组移除
- 业务层只需实现一个查询方法，不需要手动调 JoinGroup/LeaveGroup

### 内存数据结构

```
connections:  uid → []*ClientConn              // 用户连接表
groups:       groupId → map[uid]struct{}        // 群 → 本实例在线成员
userGroups:   uid → map[groupId]struct{}        // 用户 → 所属群（反向索引，断开时快速清理）
```

每个实例只存连到自己的用户的群关系，不存完整群成员列表。

### 频道隔离

- 每个群组使用独立 Redis 频道 `armylong:ws:group:{groupId}`
- 实例只在本地有该群成员时订阅，无成员时取消订阅
- `PushToGroup` 只向该群频道发消息，只有订阅了该频道的实例收到

## 代码结构

### go-library（基础设施层）

| 文件 | 职责 |
|---|---|
| `service/websocket/manager.go` | ConnManager 连接管理器 + 群组管理 + Redis Pub/Sub 桥接 |
| `service/websocket/message.go` | 统一推送消息格式 `Message{Type, Data}` |
| `service/websocket/upgrader.go` | Gin HTTP → WebSocket 升级辅助 |
| `service/websocket/instance.go` | 进程唯一 ID 生成，用于 Pub/Sub 去重 |

### armylong-go（业务接入层）

| 文件 | 职责 |
|---|---|
| `internal/websocket/handler.go` | WS 处理器：Init + GroupResolver + HandleWebSocket |
| `internal/websocket/message_types.go` | 业务消息类型常量 |
| `internal/middlewares/auth.go` | `LoadUserByToken` 导出函数，供 WS 鉴权使用 |
| `internal/controllers/routers.go` | 注册 `GET /ws` 路由 |

## 核心设计

### 1. 连接管理

- 同一用户支持多端连接（手机 + 电脑同时在线）
- `map[int64][]*ClientConn` 映射 uid 到连接列表
- 连接断开时自动清理，无剩余连接时删除 map 条目并清理群组关系
- `ClientConn.WriteJSON` 内部加写锁，gorilla/websocket 的 Conn 并发写不安全

### 2. 鉴权

- WS 连接时通过 query param `?token=xxx` 传入登录 token
- 浏览器 WebSocket API 不支持自定义 Header，只能通过 URL 参数传递
- 复用 `middlewares.LoadUserByToken` 验证身份
- 验证失败返回 401，拒绝升级

### 3. Redis Pub/Sub 跨实例推送

每个进程启动时生成唯一的 `instanceID`（16 位随机 hex）：

```
PushToUsers([1,2,3], msg)
  ├── pushLocal(1, msg)
  ├── pushLocal(2, msg)
  ├── pushLocal(3, msg)
  └── Redis PUBLISH "armylong:ws"
        {"uids": [1,2,3], "message": {...}, "from_id": "a1b2c3d4e5f6g7h8"}

PushToGroup("G1", msg)
  ├── pushLocalGroup("G1", msg) → 查本地 groups["G1"] → 遍历成员 pushLocal
  └── Redis PUBLISH "armylong:ws:group:G1"
        {"group_id": "G1", "message": {...}, "from_id": "a1b2c3d4e5f6g7h8"}

其他实例 Subscribe 收到 → from_id != 自己 → 执行推送
自己收到 → from_id == 自己 → 跳过（去重）
```

### 4. 心跳保活

- 服务端通过 `SetReadDeadline(60s)` 实现被动心跳检测
- 60 秒内无消息则断开连接，前端需实现断线重连

### 5. 消息格式

统一推送协议：

```json
{
  "type": "order_matched",
  "data": { ... }
}
```

- `type`：消息类别，决定前端怎么处理（路由到哪个处理器）
- `data`：消息内容，具体结构由 type 决定（包含 group_id 等业务字段）

消息类型定义：

| type | 触发时机 | data |
|---|---|---|
| `order_matched` | 订单匹配成功，司机接单 | 订单信息 + 行程信息 |
| `order_cancelled` | 订单被取消 | 订单ID |
| `trip_status_changed` | 行程状态变更 | 行程ID + 新状态 |
| `new_order` | 司机端收到新匹配订单 | 订单信息 |

## 业务层推送用法

```go
import (
    ws "github.com/armylong/armylong-go/internal/websocket"
    libWs "github.com/armylong/go-library/service/websocket"
)

// 单播
ws.Manager.PushToUsers([]int64{uid}, libWs.NewMessage(ws.TypeOrderMatched, orderData))

// 小群播
ws.Manager.PushToUsers([]int64{uid1, uid2, uid3}, libWs.NewMessage(ws.TypeTripStatusChanged, tripData))

// 大群播
ws.Manager.PushToGroup("G1", libWs.NewMessage(ws.TypeNewOrder, orderData))
```

## 前端接入

```javascript
const token = localStorage.getItem('token');
const ws = new WebSocket(`ws://localhost:8080/ws?token=${token}`);

ws.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    switch (msg.type) {
        case 'order_matched':
            // 处理订单匹配
            break;
        case 'trip_status_changed':
            // 处理行程状态变更
            break;
    }
};

ws.onclose = () => {
    // 指数退避重连
    setTimeout(connect, 1000);
};
```

## 水平扩展

加机器起进程即可，WebSocket 代码不需要改：

- 所有跨实例通信走 Redis Pub/Sub，实例之间无直接依赖，天然无状态
- 新实例启动后，用户连上来自动 JoinGroup + 订阅群频道
- 实例下线，连接断开，自动清理群组 + 取消订阅
- 未来瓶颈只可能在 Redis，买集群服务即可，代码层面不需要改

## 注意事项

- WebSocket 不保证消息必达，关键操作仍需 HTTP 接口确认
- 前端需实现断线重连（指数退避）
- WS 路由不走 JsonController 体系，独立处理升级
- `LoadUserByToken` 是从 `loadUserByToken` 导出的，供非中间件场景使用
- `groupResolver.GetGroups` 目前返回 nil，待群组模型建立后补充查询逻辑
