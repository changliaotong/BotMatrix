# 开发文档

## 编译问题修复记录

### 修复时间：2024-12-22

### 问题概述

BotWorker项目存在多个编译错误，主要涉及类型不匹配、函数签名问题和语法错误。

### 主要修复内容

#### 1. 类型兼容性问题

**问题描述**：`int64` 类型的用户ID与空字符串 `""` 进行比较

**修复文件**：
- `plugins/points.go` - 修复用户ID比较逻辑
- `plugins/sign_in.go` - 修复签到功能中的类型比较
- `plugins/social.go` - 修复消息处理中的类型断言

**修复方法**：
```go
// 错误代码
if userID == "" {
    return nil
}

// 正确代码
if userID == 0 {
    return nil
}
```

#### 2. Map键类型问题

**问题描述**：使用 `int64` 作为 `map[string]Type` 的键

**修复文件**：
- `plugins/points.go` - 修复积分系统的用户ID存储
- `plugins/sign_in.go` - 修复签到记录的存储

**修复方法**：
```go
// 错误代码
userPoints := p.points[userID]

// 正确代码
userIDStr := fmt.Sprintf("%d", userID)
userPoints := p.points[userIDStr]
```

#### 3. 函数签名问题

**问题描述**：函数调用参数不匹配

**修复文件**：
- `cmd/main.go` - 修复 `NewTranslatePlugin` 调用
- `plugins/points.go` - 修复 `addPoints` 函数调用
- `plugins/pets.go` - 修复 `fmt.Atoi` 调用

**修复方法**：
```go
// 错误代码
translatePlugin := plugins.NewTranslatePlugin()

// 正确代码
translatePlugin := plugins.NewTranslatePlugin(&cfg.Translate)
```

#### 4. 导入包问题

**问题描述**：使用了未导入的包或导入了未使用的包

**修复文件**：
- `plugins/pets.go` - 添加 `strconv` 导入，移除 `fmt.Atoi`
- `plugins/admin.go` - 移除未使用的 `strings` 导入
- `plugins/games.go` - 移除未使用的 `regexp` 和 `strings` 导入
- `plugins/greetings.go` - 移除未使用的 `strings` 导入
- `plugins/menu.go` - 移除未使用的 `fmt` 和 `strings` 导入
- `plugins/sign_in.go` - 移除未使用的 `strings` 导入
- `plugins/social.go` - 移除未使用的 `time` 导入

#### 5. 语法错误

**问题描述**：测试文件语法错误

**修复文件**：
- `test_cli.go` - 添加缺失的函数结束括号

**修复方法**：
```go
// 错误代码
func main() {
    // 函数体
}
func main() {
    // 另一个main函数
}

// 正确代码
func main() {
    // 函数体
}

func main() {
    // 另一个main函数
}
```

#### 6. 类型断言问题

**问题描述**：`interface{}` 类型需要类型断言

**修复文件**：
- `plugins/social.go` - 修复消息内容的类型断言

**修复方法**：
```go
// 错误代码
if strings.Contains(event.Message, "爱群主") {
    // 处理逻辑
}

// 正确代码
if msgStr, ok := event.Message.(string); ok && strings.Contains(msgStr, "爱群主") {
    // 处理逻辑
}
```

### 构建测试

修复完成后，项目可以成功构建：

```bash
# 构建成功
go build -o bot.exe ./cmd/main.go

# 测试脚本运行成功
.\test_plugins.bat
```

## 会话状态与危险操作确认（2024-12-22）

### 设计目标

- 支持多 worker、无状态部署场景
- 通过 Redis 或数据库持久化会话状态
- 为危险操作（如清空黑名单、清空白名单）提供三位数字确认码保护
- 支持多步对话 / 多级菜单流程

### 核心实现

- 在 `plugins/utils.go` 中新增全局依赖：
  - `GlobalRedis`：可选 Redis 客户端
  - `GlobalDB`：数据库连接（作为 Redis 不可用时的回退）
- 提供会话键生成工具：
  - 使用 `groupID:userID` 组合构造唯一会话键，保证同一群内按用户隔离

### 危险操作确认流程

- 使用 `StartConfirmation` 创建待确认记录：
  - 自动生成三位随机确认码和三位随机取消码
  - 默认有效期为 2 分钟
- 存储策略：
  - 若配置了 Redis：使用 `bot:confirm:<groupID>:<userID>` Key 存储 JSON
  - 否则使用数据库 `sessions` 表，`session_id` 为 `confirm:<groupID>:<userID>`
- 用户回复任意消息时，`HandleConfirmationReply` 会：
  - 从 Redis 或数据库加载待确认记录
  - 匹配确认码 / 取消码并返回统一的 `ConfirmationResult`
  - 自动清理已完成或过期的会话记录

### 多步对话 / 多级菜单

- 使用 `StartDialog` 创建对话状态：
  - 记录 `Type`、当前 `Step` 以及 `Data`（用户已输入数据）
  - 默认有效期为 5 分钟
- 状态存储：
  - Redis：Key 为 `bot:dialog:<groupID>:<userID>`
  - 数据库：`session_id` 为 `dialog:<groupID>:<userID>`
- 使用 `GetDialog` 读取当前对话状态，用 `UpdateDialog` 推进步骤：
  - 在每一步中根据 `Step` 决定提示内容和下一步逻辑
  - 完成或中断对话时调用 `EndDialog` 清理状态

### 对多 worker 的支持

- 所有与确认码、对话相关的状态都存储在 Redis 或数据库中
- 任意一个 worker 收到后续消息时：
  - 通过 `(groupID, userID)` 从共享存储恢复上下文
  - 不依赖单个进程内存，实现真正的无状态 worker

### 相关文件

- `plugins/utils.go`：会话状态、确认码、多步对话工具函数
- `plugins/moderation.go`：清空黑名单 / 白名单确认逻辑
- `plugins/dialog_demo.go`：多级菜单与多步输入示例插件

### 开发建议

1. **类型检查**：在开发过程中注意Go语言的强类型特性，确保类型匹配
2. **导入管理**：定期清理未使用的导入包
3. **代码审查**：提交代码前进行充分的测试和审查
4. **错误处理**：完善错误处理机制，提高代码健壮性

### 后续改进

- 添加更多的单元测试覆盖
- 完善错误处理机制
- 优化代码结构和性能
- 添加CI/CD自动化构建流程
