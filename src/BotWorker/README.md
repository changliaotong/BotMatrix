# BotWorker - OneBot协议兼容机器人处理程序

BotWorker是一个使用Go语言编写的兼容OneBot协议的机器人处理程序，支持WebSocket和HTTP两种通信方式，提供灵活的插件系统，方便扩展功能。

## 功能特性

- ✅ 支持OneBot v11协议
- ✅ WebSocket和HTTP双重通信支持
- ✅ 灵活的插件系统，易于扩展
- ✅ 支持私聊和群聊消息处理
- ✅ 支持各种事件类型（消息、通知、请求）
- ✅ 提供完整的API接口

### 🎯 核心功能

#### 🔍 实用工具
- **天气查询** - 实时天气信息查询
- **翻译功能** - Azure Translator API支持中英文互译
- **点歌功能** - 搜索并播放歌曲
- **报时功能** - 显示当前时间
- **计算功能** - 数学计算
- **说明书** - 插件使用说明
- **系统信息** - 服务器硬件、软件、性能信息查询

#### 🏆 成就系统
- **成就管理** - 成就解锁、进度跟踪
- **成就列表** - 查看所有可用成就
- **我的成就** - 已获得的成就
- **成就进度** - 完成进度查询
- **成就排行** - 成就排行榜

#### 🎮 游戏娱乐
- **签到系统** - 每日签到领积分
- **抽奖功能** - 随机抽奖
- **三公游戏** - 经典扑克牌游戏
- **猜拳游戏** - 石头剪刀布
- **梭哈游戏** - 经典扑克牌游戏
- **猜大小** - 骰子游戏
- **抽签解签** - 传统抽签占卜
- **运势查询** - 每日运势
- **成语接龙** - 中文成语接龙
- **笑话大全** - 随机笑话
- **鬼故事** - 恐怖故事

#### 🐾 宠物系统
- **宠物领养** - 领养可爱宠物
- **宠物喂食** - 喂养宠物
- **宠物玩耍** - 与宠物互动
- **宠物洗澡** - 清洁宠物
- **宠物升级** - 提升宠物等级
- **宠物排行** - 宠物排行榜

#### 🐎 坐骑系统
- **坐骑商店** - 购买各种坐骑
- **我的坐骑** - 查看已拥有坐骑
- **坐骑装备** - 装备坐骑
- **坐骑升级** - 提升坐骑属性
- **坐骑排行** - 坐骑排行榜

#### 🏆 积分系统
- **积分管理** - 积分获取、消耗、查询
- **打赏功能** - 给其他用户打赏积分
- **存分取分** - 积分存储和提取
- **积分榜** - 积分排名
- **买分卖分** - 积分交易
- **算力系统** - 算力获取和使用
- **余额管理** - 账户余额
- **领积分** - 每日领取积分

#### 👥 社交互动
- **早安晚安** - 问候语
- **爱群主** - 群主互动
- **头衔系统** - 用户头衔
- **变身功能** - 角色变身
- **欢迎语** - 新成员欢迎

#### 🛡️ 群管系统
- **撤回消息** - 撤回用户消息
- **禁言功能** - 禁言用户
- **踢出群聊** - 踢出用户
- **拉黑功能** - 拉黑用户
- **灰名单** - 灰名单管理
- **白名单** - 白名单管理
- **敏感词过滤** - 自动过滤敏感词
- **广告检测** - 自动检测广告
- **图片过滤** - 过滤图片消息
- **网址过滤** - 过滤网址
- **群配置** - 群管理功能配置
- **被踢加黑** - 被踢用户自动加入黑名单
- **退群加黑** - 退群用户自动加入黑名单
- **被踢提示** - 群内提示被踢事件
- **退群提示** - 群内提示退群事件

#### 🧠 智能功能
- **自动签到** - 发言自动签到
- **话唠统计** - 群内活跃度统计
- **终极智能体** - 智能对话
- **教学功能** - 机器人使用教学
- **本群信息** - 群信息查询
- **语音回复** - 将文本回复升级为 AI 语音消息（可按群开关）
- **阅后即焚** - 回复后自动撤回消息，保护聊天隐私（可按群开关）
- **多步对话 / 多级菜单** - 支持需要多次输入的信息收集与配置流程

## 项目结构

```
BotWorker/
├── cmd/
│   └── main.go              # 主程序入口
├── configs/
│   └── config.yaml          # 配置文件示例
├── docs/                    # 插件文档目录
│   ├── moderation.md        # 群管插件文档
│   ├── sign_in.md           # 签到插件文档
│   ├── translate.md         # 翻译插件文档
│   ├── music.md             # 点歌插件文档
│   ├── pets.md              # 宠物系统插件文档
│   ├── mount.md             # 坐骑系统插件文档
│   └── ...                  # 其他插件文档
├── internal/
│   ├── onebot/              # OneBot协议定义
│   │   ├── event.go         # 事件数据结构
│   │   ├── request.go       # 请求和响应数据结构
│   │   └── robot.go         # 机器人接口定义
│   ├── server/              # 服务器实现
│   │   ├── websocket.go     # WebSocket服务器
│   │   ├── http.go          # HTTP服务器
│   │   └── combined.go      # 组合服务器
│   ├── config/              # 配置管理
│   │   └── config.go        # 配置结构
│   └── utils/               # 工具函数
│       └── common.go        # 通用工具
├── plugins/                 # 插件目录
│   ├── moderation.go        # 群管插件
│   ├── sign_in.go           # 签到插件
│   ├── translate.go         # 翻译插件
│   ├── music.go             # 点歌插件
│   ├── points.go            # 积分插件
│   ├── games.go             # 游戏插件
│   ├── social.go            # 社交插件
│   ├── utils.go             # 工具插件
│   ├── admin.go             # 管理插件
│   ├── menu.go              # 菜单插件
│   ├── dialog_demo.go       # 多级菜单与多步对话示例插件
│   ├── pets.go              # 宠物系统插件
│   ├── mount.go             # 坐骑系统插件
│   └── ...                  # 其他插件
├── go.mod                   # Go模块定义
├── go.sum                   # Go模块依赖
└── README.md                # 项目文档
```

## 快速开始

### 环境要求

- Go 1.20或更高版本

### 安装依赖

```bash
go mod tidy
```

### 配置文件

复制`configs/config.yaml`并修改配置：

```yaml
# 服务器配置
server:
  websocket_port: 8080
  http_port: 8081

# 翻译插件配置
translate:
  endpoint: "https://api.cognitive.microsofttranslator.com/translate"
  api_key: "your-azure-api-key"
  timeout: 10s
  region: "eastasia"

# 天气插件配置
weather:
  api_key: "your-weather-api-key"

# 音乐插件配置
music:
  platform: "netease"
```

### 运行程序

```bash
go run cmd/main.go
```

程序将启动两个服务器：
- WebSocket服务器：`ws://localhost:8080/ws`
- HTTP服务器：`http://localhost:8081`

如需使用多 worker / 无状态部署：
- 建议在配置中启用 Redis，用于存储会话和确认状态
- 未配置 Redis 时将自动退回使用数据库存储

### 构建程序

```bash
# 构建主程序
go build -o bot.exe ./cmd/main.go

# 运行测试脚本
.\test_plugins.bat
```

## 开发说明

### 编译问题修复 (2024-12-22)

本项目近期修复了多个编译错误，主要涉及：

- **类型兼容性**：修复了 `int64` 与 `string` 类型比较的问题
- **函数签名**：修复了多个插件的函数调用参数不匹配问题
- **导入清理**：移除了未使用的导入包
- **语法错误**：修复了测试文件中的语法问题

### 插件开发

### 插件文档

## 插件开发

### 插件文档

所有插件都有详细的文档，位于`docs/`目录下。每个插件都有对应的Markdown文档，包含：
- 插件功能说明
- 命令列表
- 参数说明
- 使用示例

### 创建插件

创建一个新的插件文件，实现`plugin.Plugin`接口：

```go
package plugins

import (
    "botworker/internal/onebot"
    "botworker/internal/plugin"
    "log"
)

type MyPlugin struct{}

func (p *MyPlugin) Name() string {
    return "myplugin"
}

func (p *MyPlugin) Description() string {
    return "我的自定义插件"
}

func (p *MyPlugin) Version() string {
    return "1.0.0"
}

func (p *MyPlugin) Init(robot plugin.Robot) {
    log.Println("加载我的插件")

    // 处理消息事件
    robot.OnMessage(func(event *onebot.Event) error {
        // 处理逻辑
        return nil
    })
}
```

### 加载插件

在主程序中加载插件：

```go
// 加载自定义插件
myPlugin := &plugins.MyPlugin{}
if err := pluginManager.LoadPlugin(myPlugin); err != nil {
    log.Fatalf("加载插件失败: %v", err)
}
```

## 配置说明

配置文件使用YAML格式，示例配置位于`configs/config.yaml`：

```yaml
# WebSocket服务器配置
websocket:
  enabled: true
  address: ":8080"
  path: "/ws"

# HTTP服务器配置
http:
  enabled: true
  address: ":8081"
  event_path: "/event"
  api_path: "/api"

# 机器人配置
bot:
  name: "BotWorker"
  version: "1.0.0"
  description: "OneBot协议兼容机器人"

# 插件配置
plugins:
  enabled: true
  path: "./plugins"
```

## API接口

### 事件类型

- `message`: 消息事件
- `notice`: 通知事件
- `request`: 请求事件

### 消息类型

- `private`: 私聊消息
- `group`: 群聊消息
- `discuss`: 讨论组消息（已废弃）

### 常用API

- `send_msg`: 发送消息
- `delete_msg`: 删除消息
- `send_like`: 发送点赞
- `set_group_kick`: 踢出群成员
- `set_group_ban`: 禁言群成员

## 示例代码

### 发送私聊消息

```go
params := &onebot.SendMessageParams{
    UserID:  123456,
    Message: "你好，这是一条测试消息",
}
response, err := robot.SendMessage(params)
```

### 发送群聊消息

```go
params := &onebot.SendMessageParams{
    GroupID:  654321,
    Message: "大家好，这是一条群聊测试消息",
}
response, err := robot.SendMessage(params)
```

## 许可证

本项目使用MIT许可证，详情请查看LICENSE文件。

## 贡献

欢迎提交Issue和Pull Request来改进这个项目！

## 联系方式

如有问题或建议，请通过GitHub Issues联系我们。

## GitHub提交指南

### 提交前准备

1. **检查代码完整性**
   - 确保所有功能都已实现并测试
   - 检查所有插件文档是否完整
   - 确保配置文件示例正确

2. **更新版本信息**
   - 更新`go.mod`中的版本
   - 更新插件中的版本号

3. **创建提交说明**
   - 清晰描述本次提交的功能
   - 列出主要修改的文件
   - 说明新增的功能和改进

### 提交命令

```bash
# 初始化Git仓库（首次提交）
git init

# 添加所有文件
git add .

# 提交代码
git commit -m "feat: 完善所有功能和文档"

# 添加远程仓库
git remote add origin https://github.com/your-username/BotWorker.git

# 推送代码
git push -u origin main
```

### 提交说明示例

```
feat: 完善BotWorker机器人系统

- ✅ 实现所有核心功能（签到、积分、游戏、群管等）
- ✅ 集成Azure Translator翻译服务
- ✅ 完善群管系统（被踢加黑、退群加黑等）
- ✅ 实现自动签到功能
- ✅ 完善所有插件文档
- ✅ 更新项目README.md
- ✅ 修复已知问题
- ✅ 新增宠物系统
- ✅ 新增坐骑系统

主要文件修改：
- plugins/moderation.go - 完善群管功能
- plugins/sign_in.go - 实现自动签到
- plugins/translate.go - Azure翻译集成
- plugins/pets.go - 宠物系统
- plugins/mount.go - 坐骑系统
- docs/ - 所有插件文档
- README.md - 完善项目文档
```
