# 龙聊 - WebSocket 流程

## 一、连接建立

### 前端发起连接

```javascript
const ws = new WebSocket('ws://你的域名/ws?token=你的登录token')
```

浏览器底层会发起一个 HTTP GET 请求，携带特殊头部：

```
GET /ws?token=xxx HTTP/1.1
Upgrade: websocket
Connection: Upgrade
...
```

### 后端路由匹配

`internal/controllers/routers.go` 中注册：

```go
ws.Init()                          // 初始化连接管理器 + 订阅 Redis
engine.GET("/ws", ws.HandleWebSocket)  // 注册 WebSocket 路由
```

这是一个普通的 Gin GET 路由，跟 HTTP 接口注册方式一样。

### HandleWebSocket 处理流程

`internal/websocket/handler.go`：

```
1. ctx.Query("token") 取出 token
2. middlewares.LoadUserByToken(token) 校验 token，拿到用户信息
3. 校验失败 → 返回 401，连接拒绝
4. 校验成功 → libWs.Upgrade(ctx) 把 HTTP 连接升级为 WebSocket
5. Manager.AddConn(uid, conn) 将连接加入管理器
   - 内部自动调用 groupResolver.GetGroups(uid) 获取用户参与的所有聊天
   - 将用户加入每个 chatId 对应的频道（单聊+群聊统一）
6. 进入读取循环，SetReadDeadline 60秒无消息则断开
7. 断开时 Manager.RemoveConn(uid, client) 清理连接和频道关系
```

### 连接管理器（ConnManager）

`go-library/service/websocket/manager.go`：

- 维护 uid → []*Client 的映射（一个用户可能多个设备同时在线）
- 维护 group → map[uid] 的映射（频道关系，chatId 即频道ID）
- 通过 Redis Pub/Sub 实现跨进程消息推送

### groupResolver 的作用

groupResolver 回答一个问题：**这个用户在哪些频道里？**

用户建立 WebSocket 连接时，Manager 需要知道该把用户加入哪些频道，否则 `PushToGroup("chat_123", msg)` 推过来时，Manager 查本地 groups map 找不到 chat_123 里有谁，消息就丢了。

```
用户A连接WebSocket
  → Manager.AddConn(uid, conn)
    → 调用 groupResolver.GetGroups(uid)
    → 返回 ["chat_123", "chat_456"]
    → 把用户A加入这些频道的本地映射
    → 同时订阅这些频道的 Redis channel
```

### 频道是怎么进到实例里的

**不是 chat 主动进实例，而是用户连接时把 chat 带进实例。**

哪个实例有该 chat 的成员在线，哪个实例就订阅了该频道：

```
1. 用户A连接到实例1
2. groupResolver 返回 A 在 chat_123 里
3. 实例1 订阅 Redis 频道 "armylong:ws:group:chat_123"
4. 本地 groups map 记录: chat_123 → {A}

用户A断开连接:
  → 从 chat_123 的成员列表里移除A
  → chat_123 还有其他成员在线？保留订阅
  → chat_123 没人了？取消订阅，频道从该实例消失
```

同一个 chat 可以被多个实例同时订阅。比如 chat_123 是个 500 人大群：

```
实例1: 用户A、B在线 → 订阅了 chat_123
实例2: 用户C、D在线 → 订阅了 chat_123
实例3: 没人在线     → 没订阅
```

发消息时只有实例1和实例2收到，实例3根本收不到。

## 二、消息推送

### 推送场景

只有"别人给你发消息"才走 WebSocket 推送，因为服务端需要主动推给客户端。

> **注意**：PushToGroup 会推给频道内所有在线成员，包括发送者自己。前端通过 `from_uid === currentUid` 判断是自己发的还是对方发的。自己发的消息用于确认发送成功（去掉转圈），对方发的消息才追加到列表。

### 统一走 PushToGroup

单聊和群聊统一使用 `PushToGroup(chatId, msg)` 推送，不再使用 `PushToUsers`。

**为什么统一？**

龙聊的设计是"所有都是 chat"——单聊和群聊都是 chat 实体，只是 chat_type 不同。既然如此，推送也应该统一：

1. **语义一致**：chatId 就是频道ID，不管是单聊还是群聊，消息都属于某个 chat，推给这个 chat 频道即可
2. **精准投递**：`PushToGroup` 走独立 Redis 频道，只有该 chat 成员所在的实例收到。而 `PushToUsers` 走全局频道广播，所有实例都收到再各自检查有没有目标用户，浪费资源
3. **代码简洁**：不需要判断 chat_type 再选择推送方式，一行 `PushToGroup(chatId, msg)` 搞定

**Redis 订阅数会暴增吗？**

不会。频道是按实例订阅的，不是按用户。同一实例上 100 个用户都在同一个 chat 里，只订阅 1 个频道。Redis 处理百万级订阅毫无压力。

### 推送流程

```
用户A发消息 → HTTP POST /long_chat/message/send
  → MessageBusiness.SendMessage()
    → 1. 校验聊天状态、成员身份
    → 2. 插入消息到 tb_long_chat_message
    → 3. 更新所有参与者的 tb_long_chat_list（last_msg、unread）
    → 4. pushToMembers() 推送 WebSocket
      → 构造推送数据: {chat_id, msg_id, from_uid, msg_type, content, created_at}
      → libWs.NewMessage("long_chat", data) 构造消息
      → ws.Manager.PushToGroup(chatId, msg) 统一走频道推送
        → 本实例: 查本地 groups map，推送给该频道的在线成员
        → 其他实例: 发布到 Redis channel "armylong:ws:group:{chatId}"
```

### 跨进程推送

```
实例A: 用户A发消息 → PushToGroup("chat_123")
  → 查本地 groups map → 该 chat 在本实例的在线成员 → 直接推送
  → 发布到 Redis channel "armylong:ws:group:chat_123"

实例B: 订阅了 Redis channel "armylong:ws:group:chat_123"
  → 收到消息 → 查本地 groups map → 该 chat 在本实例的在线成员 → 直接推送
```

对比 `PushToUsers` 的全局广播：

```
PushToUsers 方案（已弃用）:
  实例A 发消息 → 发布到 Redis 全局频道 "armylong:ws"
  实例B 收到 → 解析 uid 列表 → 逐个检查本地有没有这个用户 → 没有则丢弃
  问题: 每条消息都广播给所有实例，大部分实例跟这条消息无关

PushToGroup 方案（当前）:
  实例A 发消息 → 发布到 Redis 频道 "armylong:ws:group:chat_123"
  只有订阅了该频道的实例收到 → 精准投递，无浪费
```

### 推送消息格式

```json
{
  "type": "long_chat",
  "data": {
    "chat_id": "123",
    "msg_id": "456",
    "from_uid": 1,
    "sender_name": "张三",
    "msg_type": "text",
    "content": "你好",
    "created_at": "2026-06-10T12:00:00Z"
  }
}
```

前端收到后：

```javascript
ws.on('long_chat', (data) => {
    // data.chat_id     - 聊天ID
    // data.msg_id      - 消息ID
    // data.from_uid    - 发送者uid
    // data.sender_name - 发送者昵称
    // data.msg_type    - 消息类型
    // data.content     - 消息内容
    // data.created_at  - 发送时间
});
```

## 三、HTTP 接口（拉取数据）

前端打开页面时主动请求数据，走 HTTP：

| 场景 | 方式 | 说明 |
|---|---|---|
| 打开龙聊首页，看聊天列表 | HTTP GET | 拉取 list 表数据 |
| 进入聊天详情，看历史消息 | HTTP GET | 分页拉取消息记录 |
| 打开通讯录，看好友列表 | HTTP GET | 拉取好友数据 |
| 搜索用户加好友 | HTTP POST | 搜索 + 发起请求 |
| 加好友/同意/拒绝 | HTTP POST | 操作后通过 WebSocket 通知对方 |
| 发消息 | HTTP POST | 后端处理后通过 WebSocket 推给对方 |
| 清零未读 | HTTP POST | 打开聊天详情时调用 |

## 四、完整交互流程

### 单聊流程

```
1. A 登录 → HTTP POST /auth/login → 拿到 token
2. A 建立 WebSocket → new WebSocket('/ws?token=xxx')
   → 后端 groupResolver.GetGroups(uid) 返回 A 参与的所有 chatId
   → A 自动订阅所有 chat 频道
3. A 打开龙聊首页 → HTTP POST /long_chat/chat/getChatList → 拿到聊天列表
4. A 点击 B 发消息 → HTTP POST /long_chat/message/sendMessage {chat_id, msg_type, content}
5. 后端处理:
   - 插入消息
   - 更新 A 和 B 的聊天列表（last_msg、unread）
   - PushToGroup(chatId) → B 在线则实时收到
6. B 收到 WebSocket 消息 → 更新页面
7. B 打开聊天详情 → HTTP POST /long_chat/message/listMessages → 拿历史消息
8. B 打开聊天详情时 → HTTP POST /long_chat/message/clearUnread → 清零未读
```

### 群聊流程

```
1. A 创建群 → HTTP POST /long_chat/chat/createGroup {group_name, member_uids}
2. 后端创建: chat记录 + 所有成员的member记录 + 所有成员的list记录
3. A 在群里发消息 → HTTP POST /long_chat/message/sendMessage
4. 后端处理:
   - 插入消息
   - 更新所有成员的聊天列表
   - PushToGroup(chatId) → 所有在线成员实时收到
5. 成员收到推送 → 更新页面
```

### 好友流程

```
1. A 搜索用户 → HTTP POST /long_chat/friend/search {keyword: "xxx"}
2. A 添加好友 → HTTP POST /long_chat/friend/addFriend {friend_uid}
3. 后端创建双向记录: (A, B, accepted) + (B, A, pending)
4. B 查看好友请求 → HTTP POST /long_chat/friend/listPendingRequests
5. B 同意 → HTTP POST /long_chat/friend/acceptFriend {friend_uid}
6. 后端更新: (B, A, accepted)
```

## 五、WebSocket 连接管理

### 心跳检测

- 服务端：SetReadDeadline 60秒，无消息则断开
- 前端：ws.js 内置心跳，每30秒发 ping，服务端收到任何消息（含 ping）即重置 ReadDeadline

### 断线重连

- 前端 ws.js 内置自动重连，指数退避（1s → 2s → 4s → 8s → 最大30s）
- 重连时重新带上 token 鉴权
- 重连成功后前端应重新拉取聊天列表，补齐离线期间的消息

### 多设备

- 同一 uid 可以有多个 WebSocket 连接（手机+电脑）
- Manager 维护 uid → []*Client 映射
- 推送时遍历该 uid 的所有连接，全部推送
