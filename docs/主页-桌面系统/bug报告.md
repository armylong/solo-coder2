# 主页-桌面系统 Bug 报告

## 一、严重级别定义

| 级别 | 说明 |
|------|------|
| P0 - 致命 | 系统崩溃、数据丢失、安全漏洞 |
| P1 - 高危 | 主要功能失效、用户体验严重受损 |
| P2 - 中危 | 次要功能问题、体验不佳 |
| P3 - 低危 | 轻微问题、建议优化 |

---

## 二、Bug 列表

### Bug #1: 后端错误被静默忽略

**严重级别**: P1 - 高危

**问题描述**:
后端代码中大量使用 `_ =` 忽略错误返回值，导致系统问题无法被发现和追踪。

**影响文件**:
- `internal/model/desktop/tb_app.go`
- `internal/model/desktop/tb_desktop_os_app.go`
- `internal/business/index/home.go`

**问题代码示例**:

```go
// tb_app.go
func init() {
    _ = TbAppModel.CreateTable()      // 忽略错误
    _ = TbAppModel.InitDefaultApps()  // 忽略错误
}

// tb_app.go - InitDefaultApps 方法
func (m *tbAppModel) InitDefaultApps() error {
    for _, app := range defaultApps() {
        existing, err := m.GetByAppName(app.AppName)
        if err != nil || existing == nil {
            _, _ = m.Create(app)  // 忽略错误
        } else {
            if existing.Permission != app.Permission {
                existing.Permission = app.Permission
                _ = m.Update(existing)  // 忽略错误
            }
        }
    }
    return nil
}

// home.go - installLongStoreForUser 方法
func (h *homeBusiness) installLongStoreForUser(uid, longStoreAppId int64) {
    // ...
    _ = desktopModel.TbUserAppModel.CreateOrUpdate(  // 忽略错误
        uid,
        longStoreAppId,
        ext,
        desktopModel.UserAppStatusInstalled,
    )
}
```

**影响**:
1. 系统初始化失败时无法发现
2. 数据库操作错误无法追踪
3. 业务逻辑错误无反馈
4. 线上问题排查困难

**复现步骤**:
1. 故意破坏数据库连接
2. 启动系统
3. 观察到无任何错误提示，但功能异常

**修复建议**:
1. 添加日志记录，至少使用 `log.Printf` 记录错误
2. 关键操作的错误应该向上层返回
3. 初始化阶段的错误应该 panic 或记录到日志

---

### Bug #2: 数据模型设计冗余

**严重级别**: P2 - 中危

**问题描述**:
存在两张应用表，职责不清，可能导致数据同步问题和维护混乱。

**影响文件**:
- `internal/model/desktop/tb_app.go`
- `internal/model/desktop/tb_desktop_os_app.go`

**问题分析**:

| 表名 | 用途 | 实际使用情况 |
|------|------|--------------|
| `tb_app` | 系统应用表 | 主要使用，Long Store 在此表 |
| `tb_desktop_os_app` | 桌面系统应用表 | 定义了 12 个默认应用，但代码中未使用 |

**代码对比**:

```go
// tb_app.go - 默认应用只有 Long Store
func defaultApps() []*TbApp {
    return []*TbApp{
        {AppName: LongStoreAppName, Desc: "应用商店...", Icon: "🏪", ...},
    }
}

// tb_desktop_os_app.go - 默认应用有 12 个
func (m *tbDesktopOsAppModel) InitDefaultApps() error {
    defaultApps := []*TbDesktopOsApp{
        {AppName: "gomoku", Label: "五子棋", ...},
        {AppName: "chinese-chess", Label: "中国象棋", ...},
        {AppName: "go", Label: "围棋", ...},
        // ... 共 12 个应用
    }
    // ...
}
```

**影响**:
1. 新开发者困惑该使用哪张表
2. 两张表数据可能不同步
3. 维护成本增加
4. 潜在的逻辑混乱

**修复建议**:
1. 明确两张表的职责，或合并为一张表
2. 如果 `tb_desktop_os_app` 是废弃的，应该标注并逐步移除
3. 统一应用管理逻辑

---

### Bug #3: 前端坐标值无边界检查

**严重级别**: P2 - 中危

**问题描述**:
前端保存应用坐标时，没有进行边界检查，可能导致图标超出可视范围。

**影响文件**:
- `static/index/src/main.js`

**问题代码**:

```javascript
// handleDrop 函数 - 计算坐标后直接保存
function handleDrop(e) {
    // ...
    // 桌面 → 桌面（移动位置）
    } else if (state.dragData.type === 'desktop' && !isDroppingToDock) {
        const appData = state.desktopApps[state.dragData.index];
        // 直接计算，无边界检查
        const newX = Math.round(((e.clientX - desktopRect.left) / desktopRect.width) * 100);
        const newY = Math.round(((e.clientY - desktopRect.top - 25) / (desktopRect.height - 80)) * 100);
        appData.x = newX;
        appData.y = newY;
        // 直接保存，可能超出 0-100 范围
        await saveDesktopApp(appData.app_id, newX, newY);
        renderDesktop();
    }
    // ...
}
```

**问题场景**:
1. 用户将图标拖到桌面区域外
2. `e.clientX` 或 `e.clientY` 超出桌面边界
3. 计算出的 `newX` 或 `newY` 可能小于 0 或大于 100
4. 保存后，下次加载时图标位置异常

**影响**:
1. 图标可能超出可视范围
2. 用户无法看到或操作该图标
3. 界面显示异常

**复现步骤**:
1. 登录系统
2. 尝试将图标拖到桌面边缘外
3. 刷新页面
4. 观察图标位置是否异常

**修复建议**:
1. 添加边界检查，确保坐标在合理范围内
2. 建议使用 `Math.max(0, Math.min(100, value))` 进行限制
3. 后端也应该进行验证

---

### Bug #4: 前端状态管理可能导致竞态条件

**严重级别**: P2 - 中危

**问题描述**:
前端使用单一全局 `state` 对象管理状态，异步操作可能导致竞态条件。

**影响文件**:
- `static/index/src/main.js`

**问题场景**:

```javascript
// 场景 1: 快速拖拽多个图标
// 用户快速拖拽图标 A 和图标 B
// 两个 saveDesktopApp 请求几乎同时发出
// 后返回的请求可能覆盖先返回的状态

// 场景 2: 多标签页操作
// 用户在两个标签页登录同一账号
// 在标签页 A 移动图标
// 在标签页 B 移动同一图标
// 数据可能不一致

// 场景 3: 网络延迟导致请求乱序
async function handleDrop(e) {
    // ...
    // 假设网络延迟，请求 1 比请求 2 晚返回
    await saveDesktopApp(appData.app_id, newX, newY);  // 请求 1
    renderDesktop();  // 用请求 1 的结果渲染
}

// 另一个操作
async function someOtherOperation() {
    await saveDesktopApp(appData.app_id, otherX, otherY);  // 请求 2
    renderDesktop();  // 用请求 2 的结果渲染
}
```

**影响**:
1. 状态不一致
2. 数据可能被错误覆盖
3. 用户操作无反馈或反馈错误

**修复建议**:
1. 为状态操作添加请求队列或乐观锁
2. 使用版本号或时间戳检测冲突
3. 关键操作添加 Loading 状态，防止重复操作

---

### Bug #5: 错误提示不友好

**严重级别**: P3 - 低危

**问题描述**:
前端使用原生 `alert()` 显示错误信息，用户体验差且不专业。

**影响文件**:
- `static/index/src/main.js`

**问题代码**:

```javascript
// 使用原生 alert
case 'settings':
    alert('系统设置功能开发中...');
    break;

// 错误提示
catch (e) {
    console.error('Uninstall failed:', e);
    alert('卸载失败: ' + e.message);  // 显示技术化错误信息
}
```

**影响**:
1. 用户体验差
2. 错误信息技术化，用户难以理解
3. 缺少错误恢复引导
4. 缺乏专业感

**修复建议**:
1. 使用自定义的 Toast 或 Modal 组件
2. 错误信息应该用户友好，避免技术术语
3. 提供错误恢复建议（如"请重试"、"检查网络连接"）
4. 记录详细错误到控制台或日志系统

---

### Bug #6: 缺少输入验证

**严重级别**: P2 - 中危

**问题描述**:
前端和后端对输入参数缺少验证，可能导致异常数据。

**影响范围**:
- 登录/注册表单
- 坐标保存
- 应用 ID 等参数

**潜在问题**:

```javascript
// 登录表单 - 缺少长度、格式验证
const data = {};
for (const [key, value] of formData.entries()) {
    data[key] = value;  // 直接使用，无验证
}

// 坐标保存 - 后端缺少验证
// 后端接收的 x, y 可能是负数或超大数
```

**影响**:
1. 异常数据可能导致系统崩溃
2. 安全隐患（SQL 注入、XSS 等）
3. 数据一致性问题

**修复建议**:
1. 前端添加表单验证（长度、格式、必填等）
2. 后端添加参数校验
3. 使用参数化查询防止 SQL 注入
4. 对特殊字符进行转义

---

### Bug #7: 前端代码函数过长

**严重级别**: P3 - 低危

**问题描述**:
部分函数代码量过大，职责不单一，难以维护和测试。

**影响文件**:
- `static/index/src/main.js`

**问题分析**:

| 函数 | 预估行数 | 问题 |
|------|----------|------|
| `render()` | ~100+ 行 | 包含大量 HTML 模板字符串 |
| `handleDrop()` | ~80 行 | 处理多种拖拽场景 |
| `handleMenuAction()` | ~50 行 | 处理多种菜单操作 |

**问题代码示例**:

```javascript
// render 函数包含大量模板
function render() {
    const app = document.getElementById('app');
    app.innerHTML = `
        <!-- 超过 100 行的 HTML 模板 -->
        <div class="statusbar">
            <div class="statusbar-left">
                <div class="apple-logo">🍎</div>
                <div class="site-name">阿米龙</div>
            </div>
            <!-- ... 更多内容 -->
        </div>
        <div class="apple-menu">...</div>
        <div class="desktop">...</div>
        <div class="dock">...</div>
        <div class="context-menu">...</div>
        <div class="about-modal">...</div>
        <div class="login-modal">...</div>
    `;
    renderDesktop();
    renderDock();
}
```

**影响**:
1. 可读性差
2. 难以单元测试
3. 修改风险高
4. 维护困难

**修复建议**:
1. 将大函数拆分为多个小函数
2. 模板字符串可以提取为独立的模板函数
3. 使用事件委托减少重复代码
4. 考虑使用组件化思想重构

---

### Bug #8: 缺少并发安全保护

**严重级别**: P2 - 中危

**问题描述**:
系统缺少并发安全保护，多用户或多标签页操作可能导致数据不一致。

**问题场景**:

1. **多标签页操作**:
   - 用户在标签页 A 打开应用
   - 用户在标签页 B 卸载该应用
   - 标签页 A 的窗口状态异常

2. **快速连续操作**:
   - 用户快速点击多个图标
   - 多个异步请求同时发出
   - 状态可能被错误覆盖

3. **后端并发**:
   - 多个请求同时修改同一用户数据
   - 缺少数据库事务
   - 可能导致数据不一致

**影响**:
1. 数据不一致
2. 界面显示异常
3. 用户操作无反馈

**修复建议**:
1. 前端使用请求队列或乐观锁
2. 关键数据库操作使用事务
3. 添加版本号或时间戳检测冲突
4. 使用 WebSocket 或轮询同步多标签页状态

---

## 三、Bug 统计

### 按严重级别统计

| 级别 | 数量 | 百分比 |
|------|------|--------|
| P1 - 高危 | 1 | 12.5% |
| P2 - 中危 | 4 | 50.0% |
| P3 - 低危 | 3 | 37.5% |
| **总计** | **8** | **100%** |

### 按模块统计

| 模块 | Bug 数量 |
|------|----------|
| 后端错误处理 | 1 |
| 数据模型 | 1 |
| 前端状态管理 | 2 |
| 输入验证 | 1 |
| 用户体验 | 1 |
| 代码质量 | 1 |
| 并发安全 | 1 |

---

## 四、修复优先级建议

### 第一优先级（立即修复）

1. **Bug #1: 后端错误被静默忽略**
   - 影响：系统问题无法发现
   - 建议：至少添加日志记录

### 第二优先级（本周修复）

2. **Bug #3: 前端坐标值无边界检查**
3. **Bug #6: 缺少输入验证**
4. **Bug #4: 前端状态管理可能导致竞态条件**
5. **Bug #8: 缺少并发安全保护**

### 第三优先级（迭代优化）

6. **Bug #2: 数据模型设计冗余**
7. **Bug #5: 错误提示不友好**
8. **Bug #7: 前端代码函数过长**

---

## 五、测试建议

### 建议添加的测试用例

1. **边界测试**:
   - 坐标值边界测试（0, 100, -1, 101）
   - 空值、异常值输入测试

2. **并发测试**:
   - 多标签页同时操作测试
   - 快速连续操作测试

3. **错误处理测试**:
   - 数据库连接失败测试
   - 网络异常测试
   - 权限不足测试

4. **兼容性测试**:
   - 不同浏览器测试
   - 移动端测试
   - 低性能设备测试

### 建议添加的自动化测试

1. **后端单元测试**:
   - 模型层测试
   - 业务逻辑测试
   - 错误处理测试

2. **前端单元测试**:
   - 状态管理测试
   - 事件处理测试
   - 工具函数测试

3. **集成测试**:
   - API 接口测试
   - 端到端流程测试
