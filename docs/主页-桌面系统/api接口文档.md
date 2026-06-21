# 主页-桌面系统 API 接口文档

## 一、接口概述

### 1.1 基础信息

- **协议**: HTTP/HTTPS
- **数据格式**: JSON
- **编码**: UTF-8
- **认证方式**: Bearer Token (Authorization Header)

### 1.2 通用响应格式

所有接口返回统一的 JSON 格式：

```json
{
    "code": 0,
    "message": "success",
    "data": {
        // 业务数据
    }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| code | int | 状态码，>=0 表示成功（0 为通用成功），<0 表示失败（-1 为通用失败） |
| message | string | 状态信息，成功时为 "success" |
| data | object/array | 响应数据，具体内容由各接口定义 |

### 1.3 通用错误码

| 状态码 | 说明 |
|--------|------|
| 0 | 成功 |
| -401 | 未授权/登录失效 |
| -1 | 通用失败 |
| -500 | 服务器内部错误 |

### 1.4 认证方式

需要登录的接口需要在请求头中携带认证令牌：

```
Authorization: <token>
```

**说明**:
- Token 在登录接口返回
- Token 存储在前端 localStorage 中
- 过期或无效的 Token 会返回 -401 错误

---

## 二、桌面系统接口

### 2.1 获取桌面数据

#### 接口信息

| 项目 | 说明 |
|------|------|
| 接口路径 | `/index/desktopOs` |
| 请求方法 | GET |
| 是否需要登录 | 是 |

#### 功能描述

获取当前登录用户的桌面系统数据，包括：
- 用户信息
- 用户权限级别
- 已安装应用列表
- 桌面和 Dock 布局信息

#### 请求参数

无查询参数，通过 Authorization 头识别用户。

#### 请求示例

```javascript
// 使用 fetch 调用
const response = await fetch('/index/desktopOs', {
    headers: {
        'Authorization': 'your-token-here'
    }
});
const result = await response.json();
```

#### 响应示例

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "user": {
            "uid": 1,
            "account": "admin",
            "name": "管理员",
            "password": "",
            "status": 1,
            "created_at": "2024-01-01T00:00:00Z",
            "updated_at": "2024-01-01T00:00:00Z"
        },
        "user_permission": 1,
        "apps": [
            {
                "app_id": 1,
                "app_name": "Long Store",
                "desc": "应用商店，浏览和安装各类应用",
                "icon": "🏪",
                "url": "long-store/index.html",
                "type": 1,
                "permission": 0,
                "status": 1,
                "created_at": "2024-01-01T00:00:00Z",
                "updated_at": "2024-01-01T00:00:00Z"
            }
        ],
        "layout": {
            "desktop_apps": [
                {
                    "app_id": 1,
                    "app_name": "Long Store",
                    "x": 2,
                    "y": 2
                }
            ],
            "dock_apps": []
        }
    }
}
```

#### 响应字段说明

**data**:

| 字段 | 类型 | 说明 |
|------|------|------|
| user | object | 用户信息对象 |
| user_permission | int | 用户权限级别（0=普通用户，1=超管，2=管理员） |
| apps | array | 用户可见且已安装的应用列表 |
| layout | object | 桌面布局信息 |

**layout 对象**:

| 字段 | 类型 | 说明 |
|------|------|------|
| desktop_apps | array | 桌面应用列表 |
| dock_apps | array | Dock 应用列表 |

**desktop_apps 元素**:

| 字段 | 类型 | 说明 |
|------|------|------|
| app_id | int64 | 应用 ID |
| app_name | string | 应用名称 |
| x | int | 桌面 X 坐标（百分比） |
| y | int | 桌面 Y 坐标（百分比） |

**dock_apps 元素**:

| 字段 | 类型 | 说明 |
|------|------|------|
| app_id | int64 | 应用 ID |
| app_name | string | 应用名称 |
| dock_index | int | Dock 栏索引位置 |

**app 元素**:

| 字段 | 类型 | 说明 |
|------|------|------|
| app_id | int64 | 应用 ID |
| app_name | string | 应用名称 |
| desc | string | 应用描述 |
| icon | string | 应用图标（emoji） |
| url | string | 应用入口 URL |
| type | int | 应用类型（1=应用，2=游戏） |
| permission | int | 应用权限级别 |
| status | int | 状态（1=启用） |

#### 业务逻辑说明

1. **权限过滤**: 根据用户权限级别返回对应的可见应用
2. **自动安装 Long Store**: 如果用户未安装 Long Store，自动为其安装
3. **布局构建**: 从 `tb_user_app` 表读取用户应用的位置信息

#### 前端调用位置

`static/index/src/main.js` 中的 `loadDesktopOs()` 函数

---

## 三、设置接口

### 3.1 保存桌面应用位置

#### 接口信息

| 项目 | 说明 |
|------|------|
| 接口路径 | `/settings/setDesktopApp` |
| 请求方法 | POST |
| 是否需要登录 | 是 |
| Content-Type | application/json |

#### 功能描述

保存应用在桌面上的位置坐标。

#### 请求参数

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| uid | int64 | 是 | 用户 ID |
| app_id | int64 | 是 | 应用 ID |
| x | int | 是 | 桌面 X 坐标（百分比，0-100） |
| y | int | 是 | 桌面 Y 坐标（百分比，0-100） |

#### 请求示例

```javascript
const response = await fetch('/settings/setDesktopApp', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json',
        'Authorization': 'your-token-here'
    },
    body: JSON.stringify({
        uid: 1,
        app_id: 1,
        x: 12,
        y: 2
    })
});
const result = await response.json();
```

#### 响应示例

```json
{
    "code": 0,
    "message": "success",
    "data": null
}
```

#### 响应字段说明

无额外响应数据，code >= 0 即表示成功。

#### 业务逻辑说明

1. **更新用户应用记录**: 更新 `tb_user_app` 表中对应记录的 `ext` 字段
2. **设置位置类型**: 将 `ext.position` 设置为 `"desktop"`
3. **保存坐标**: 将 `ext.x` 和 `ext.y` 设置为传入的坐标值

#### 前端调用位置

`static/index/src/main.js` 中的 `saveDesktopApp()` 函数

---

### 3.2 保存 Dock 应用位置

#### 接口信息

| 项目 | 说明 |
|------|------|
| 接口路径 | `/settings/setDockApp` |
| 请求方法 | POST |
| 是否需要登录 | 是 |
| Content-Type | application/json |

#### 功能描述

保存应用在 Dock 栏中的位置索引。

#### 请求参数

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| uid | int64 | 是 | 用户 ID |
| app_id | int64 | 是 | 应用 ID |
| dock_index | int | 是 | Dock 栏索引位置（从 0 开始） |

#### 请求示例

```javascript
const response = await fetch('/settings/setDockApp', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json',
        'Authorization': 'your-token-here'
    },
    body: JSON.stringify({
        uid: 1,
        app_id: 1,
        dock_index: 0
    })
});
const result = await response.json();
```

#### 响应示例

```json
{
    "code": 0,
    "message": "success",
    "data": null
}
```

#### 响应字段说明

无额外响应数据，code >= 0 即表示成功。

#### 业务逻辑说明

1. **更新用户应用记录**: 更新 `tb_user_app` 表中对应记录的 `ext` 字段
2. **设置位置类型**: 将 `ext.position` 设置为 `"dock"`
3. **保存索引**: 将 `ext.dock_index` 设置为传入的索引值

#### 前端调用位置

`static/index/src/main.js` 中的 `saveDockApp()` 函数

---

## 四、应用商店接口

### 4.1 卸载应用

#### 接口信息

| 项目 | 说明 |
|------|------|
| 接口路径 | `/long_store/uninstall` |
| 请求方法 | POST |
| 是否需要登录 | 是 |
| Content-Type | application/json |

#### 功能描述

卸载用户已安装的应用（软删除）。

#### 请求参数

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| app_id | int64 | 是 | 要卸载的应用 ID |

#### 请求示例

```javascript
const response = await fetch('/long_store/uninstall', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json',
        'Authorization': 'your-token-here'
    },
    body: JSON.stringify({
        app_id: 2
    })
});
const result = await response.json();
```

#### 响应示例

```json
{
    "code": 0,
    "message": "success",
    "data": null
}
```

#### 响应字段说明

无额外响应数据，code >= 0 即表示成功。

#### 业务逻辑说明

1. **软删除**: 更新 `tb_user_app` 表中记录的 `status` 字段为 `2`（已卸载）
2. **保留记录**: 不物理删除记录，保留操作历史
3. **权限检查**: 只允许卸载当前登录用户的应用

#### 前端调用位置

`static/index/src/main.js` 中的 `handleUninstall()` 函数

---

## 五、认证接口

### 5.1 用户登录

#### 接口信息

| 项目 | 说明 |
|------|------|
| 接口路径 | `/auth/login` |
| 请求方法 | POST |
| 是否需要登录 | 否 |
| Content-Type | application/json |

#### 功能描述

用户登录认证，返回认证令牌。

#### 请求参数

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| account | string | 是 | 用户账号 |
| password | string | 是 | 用户密码 |
| device_type | string | 否 | 设备类型，默认 "pc" |

#### 请求示例

```javascript
const response = await fetch('/auth/login', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json'
    },
    body: JSON.stringify({
        account: 'admin',
        password: '123456',
        device_type: 'pc'
    })
});
const result = await response.json();
```

#### 响应示例

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "user": {
            "uid": 1,
            "account": "admin",
            "name": "管理员",
            "status": 1,
            "created_at": "2024-01-01T00:00:00Z",
            "updated_at": "2024-01-01T00:00:00Z"
        },
        "token": "eyJhbGciOiJIUzI1NiIs..."
    }
}
```

#### 响应字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| user | object | 用户信息对象 |
| token | string | 认证令牌，用于后续接口调用 |

#### 业务逻辑说明

1. **账号密码验证**: 验证账号和密码是否匹配
2. **状态检查**: 检查用户状态是否正常
3. **生成令牌**: 生成 JWT 或其他格式的认证令牌
4. **记录登录**: 可能记录登录日志

#### 前端调用位置

`static/index/src/main.js` 中的 `handleLoginSubmit()` 函数

---

### 5.2 用户注册

#### 接口信息

| 项目 | 说明 |
|------|------|
| 接口路径 | `/auth/register` |
| 请求方法 | POST |
| 是否需要登录 | 否 |
| Content-Type | application/json |

#### 功能描述

创建新用户账号。

#### 请求参数

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| account | string | 是 | 用户账号（唯一） |
| name | string | 是 | 用户昵称/名称 |
| password | string | 是 | 用户密码 |
| device_type | string | 否 | 设备类型，默认 "pc" |

#### 请求示例

```javascript
const response = await fetch('/auth/register', {
    method: 'POST',
    headers: {
        'Content-Type': 'application/json'
    },
    body: JSON.stringify({
        account: 'newuser',
        name: '新用户',
        password: '123456',
        device_type: 'pc'
    })
});
const result = await response.json();
```

#### 响应示例

```json
{
    "code": 0,
    "message": "success",
    "data": {
        "user": {
            "uid": 2,
            "account": "newuser",
            "name": "新用户",
            "status": 1,
            "created_at": "2024-01-01T00:00:00Z",
            "updated_at": "2024-01-01T00:00:00Z"
        },
        "token": "eyJhbGciOiJIUzI1NiIs..."
    }
}
```

#### 响应字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| user | object | 新创建的用户信息 |
| token | string | 认证令牌（注册成功后自动登录） |

#### 业务逻辑说明

1. **账号唯一性检查**: 检查账号是否已存在
2. **密码加密**: 对密码进行哈希加密存储
3. **创建用户**: 在用户表中创建新记录
4. **自动登录**: 注册成功后返回认证令牌

#### 前端调用位置

`static/index/src/main.js` 中的 `handleLoginSubmit()` 函数（当 loginTab 为 'register' 时）

---

### 5.3 用户登出

#### 接口信息

| 项目 | 说明 |
|------|------|
| 接口路径 | `/auth/logout` |
| 请求方法 | POST |
| 是否需要登录 | 是 |

#### 功能描述

用户登出，使当前令牌失效。

#### 请求参数

无请求体参数，通过 Authorization 头识别用户。

#### 请求示例

```javascript
const response = await fetch('/auth/logout', {
    method: 'POST',
    headers: {
        'Authorization': 'your-token-here'
    }
});
const result = await response.json();
```

#### 响应示例

```json
{
    "code": 0,
    "message": "success",
    "data": null
}
```

#### 响应字段说明

无额外响应数据，code >= 0 即表示成功。

#### 业务逻辑说明

1. **令牌失效**: 可能使当前令牌失效（取决于实现）
2. **记录登出**: 可能记录登出日志

#### 前端处理

前端在调用登出接口后需要：
1. 清除 localStorage 中的 token
2. 重置前端状态
3. 重新渲染界面

#### 前端调用位置

`static/index/src/main.js` 中的 `handleLogout()` 函数

---

## 六、接口调用流程图

### 6.1 完整登录流程

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│  前端    │     │  登录接口 │     │  桌面接口 │     │  数据库  │
└────┬─────┘     └────┬─────┘     └────┬─────┘     └────┬─────┘
     │                │                │                │
     │  POST /auth/login              │                │
     │───────────────►│                │                │
     │                │                │                │
     │                │  验证账号密码  │                │
     │                │───────────────►│                │
     │                │                │                │
     │                │◄───────────────│                │
     │                │  返回 token    │                │
     │◄───────────────│                │                │
     │                │                │                │
     │  保存 token 到 localStorage    │                │
     │                │                │                │
     │  GET /index/desktopOs          │                │
     │───────────────────────────────►│                │
     │                │                │                │
     │                │                │  查询用户信息   │
     │                │                │───────────────►│
     │                │                │                │
     │                │                │  查询应用列表   │
     │                │                │───────────────►│
     │                │                │                │
     │                │                │  查询用户布局   │
     │                │                │───────────────►│
     │                │                │                │
     │◄───────────────────────────────│                │
     │  返回桌面数据                    │                │
     │                │                │                │
     │  渲染桌面和 Dock                │                │
     │                │                │                │
```

### 6.2 拖拽保存位置流程

```
┌──────────┐     ┌──────────┐     ┌──────────┐
│  前端    │     │ 设置接口  │     │  数据库  │
└────┬─────┘     └────┬─────┘     └────┬─────┘
     │                │                │
     │  用户拖拽图标   │                │
     │                │                │
     │  计算新坐标     │                │
     │                │                │
     │  POST /settings/setDesktopApp  │
     │───────────────►│                │
     │                │                │
     │                │  更新用户应用记录│
     │                │───────────────►│
     │                │                │
     │◄───────────────│                │
     │  返回成功       │                │
     │                │                │
     │  更新本地状态   │                │
     │  重新渲染       │                │
     │                │                │
```

---

## 七、前端接口封装

### 7.1 fetchWithAuth 封装

前端统一使用 `fetchWithAuth` 函数调用接口，自动添加 Authorization 头：

```javascript
function fetchWithAuth(url, options = {}) {
    const headers = options.headers || {};
    if (state.token) {
        headers['Authorization'] = state.token;
    }
    return fetch(url, { ...options, headers });
}
```

### 7.2 响应处理模式

前端统一的响应处理模式：

```javascript
try {
    const result = await fetchWithAuth(url, options);
    const data = await result.json();
    
    if (data.code >= 0) {
        // 处理成功响应，业务数据在 data.data 中
    } else if (data.code === -401) {
        // 处理认证失败，清除 token
        saveTokenToStorage(null);
    } else {
        // 处理业务错误
        console.error(data.message);
    }
} catch (e) {
    // 处理网络错误
    console.error('Request failed:', e);
}
```

---

## 八、注意事项

### 8.1 安全注意

1. **Token 安全**: Token 应妥善保管，避免泄露
2. **HTTPS**: 生产环境应使用 HTTPS 协议
3. **输入验证**: 前端和后端都应对输入参数进行验证
4. **权限检查**: 后端应对所有需要登录的接口进行权限验证

### 8.2 性能注意

1. **减少请求**: 合并相关接口，减少网络请求
2. **缓存策略**: 对不频繁变更的数据进行缓存
3. **错误重试**: 网络错误时应有重试机制
4. **Loading 状态**: 请求期间显示 Loading 状态

### 8.3 兼容性注意

1. **浏览器兼容**: 注意 `backdrop-filter` 等 CSS 属性的浏览器兼容性
2. **Fetch API**: 旧版浏览器可能需要 polyfill
3. **localStorage**: 注意存储容量限制
4. **全屏 API**: 不同浏览器的全屏 API 实现有差异

---

## 九、接口速查表

| 接口路径 | 方法 | 功能 | 登录 |
|----------|------|------|------|
| `/index/desktopOs` | GET | 获取桌面数据 | 是 |
| `/settings/setDesktopApp` | POST | 保存桌面应用位置 | 是 |
| `/settings/setDockApp` | POST | 保存 Dock 应用位置 | 是 |
| `/long_store/uninstall` | POST | 卸载应用 | 是 |
| `/auth/login` | POST | 用户登录 | 否 |
| `/auth/register` | POST | 用户注册 | 否 |
| `/auth/logout` | POST | 用户登出 | 是 |
