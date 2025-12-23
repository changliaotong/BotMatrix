# 宝宝系统与结婚系统文档

## 1. 系统概述

BotMatrix机器人实现了两个核心功能系统：**宝宝系统**和**结婚系统**。这两个系统旨在增强机器人与用户之间的互动性，提供丰富的虚拟社交体验。

### 1.1 宝宝系统
宝宝系统允许用户拥有并培养虚拟宝宝，包括宝宝降临、学习、打工、互动等功能，通过成长值系统记录宝宝的成长过程。

### 1.2 结婚系统
结婚系统允许用户在虚拟世界中寻找伴侣，进行求婚、结婚、离婚等操作，并提供喜糖、红包、甜蜜爱心等互动元素。

## 2. 数据模型

### 2.1 宝宝系统数据模型

#### Baby 宝宝基本信息
```go
type Baby struct {
    ID          uint      `gorm:"primaryKey" json:"id"`
    UserID      string    `gorm:"size:20;index" json:"user_id"`
    Name        string    `gorm:"size:50" json:"name"`
    Birthday    time.Time `json:"birthday"`
    GrowthValue int       `json:"growth_value"`
    DaysOld     int       `json:"days_old"`
    Level       int       `json:"level"`
    Status      string    `gorm:"size:20;default:active" json:"status"` // active, abandoned
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

#### BabyEvent 宝宝事件记录
```go
type BabyEvent struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    BabyID    uint      `json:"baby_id"`
    EventType string    `gorm:"size:50" json:"event_type"` // birthday, learn, work, interact
    Content   string    `gorm:"size:255" json:"content"`
    CreatedAt time.Time `json:"created_at"`
}
```

#### BabyConfig 宝宝系统配置
```go
type BabyConfig struct {
    ID          uint      `gorm:"primaryKey" json:"id"`
    IsEnabled   bool      `gorm:"default:true" json:"is_enabled"`
    GrowthRate  int       `gorm:"default:1000" json:"growth_rate"` // 每1000成长值增加1天
    UpdateAt    time.Time `json:"update_at"`
}
```

### 2.2 结婚系统数据模型

#### UserMarriage 用户婚姻信息
```go
type UserMarriage struct {
    ID              uint      `gorm:"primaryKey" json:"id"`
    UserID          string    `gorm:"size:20;index" json:"user_id"`
    SpouseID        string    `gorm:"size:20;index" json:"spouse_id"`
    MarriageDate    time.Time `json:"marriage_date"`
    DivorceDate     time.Time `json:"divorce_date"`
    Status          string    `gorm:"size:20;default:single" json:"status"` // single, married, divorced
    SweetsCount     int       `gorm:"default:0" json:"sweets_count"`
    RedPacketsCount int       `gorm:"default:0" json:"red_packets_count"`
    SweetHearts     int       `gorm:"default:0" json:"sweet_hearts"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}
```

#### MarriageProposal 求婚记录
```go
type MarriageProposal struct {
    ID           uint      `gorm:"primaryKey" json:"id"`
    ProposerID   string    `gorm:"size:20;index" json:"proposer_id"`
    RecipientID  string    `gorm:"size:20;index" json:"recipient_id"`
    Status       string    `gorm:"size:20;default:pending" json:"status"` // pending, accepted, rejected
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
```

#### WeddingItem 婚礼物品
```go
type WeddingItem struct {
    ID          uint      `gorm:"primaryKey" json:"id"`
    ItemType    string    `gorm:"size:20" json:"item_type"` // dress, ring
    Name        string    `gorm:"size:50" json:"name"`
    Price       int       `gorm:"default:0" json:"price"`
    Description string    `gorm:"size:255" json:"description"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

#### Sweets 喜糖记录
```go
type Sweets struct {
    ID          uint      `gorm:"primaryKey" json:"id"`
    UserID      string    `gorm:"size:20;index" json:"user_id"`
    Amount      int       `json:"amount"`
    Type        string    `gorm:"size:20" json:"type"` // send, receive, eat
    Description string    `gorm:"size:255" json:"description"`
    CreatedAt   time.Time `json:"created_at"`
}
```

#### RedPacket 红包记录
```go
type RedPacket struct {
    ID          uint      `gorm:"primaryKey" json:"id"`
    UserID      string    `gorm:"size:20;index" json:"user_id"`
    Amount      int       `json:"amount"`
    Type        string    `gorm:"size:20" json:"type"` // send, receive
    Description string    `gorm:"size:255" json:"description"`
    CreatedAt   time.Time `json:"created_at"`
}
```

#### SweetHeart 甜蜜爱心
```go
type SweetHeart struct {
    ID          uint      `gorm:"primaryKey" json:"id"`
    SenderID    string    `gorm:"size:20;index" json:"sender_id"`
    RecipientID string    `gorm:"size:20;index" json:"recipient_id"`
    Amount      int       `json:"amount"`
    CreatedAt   time.Time `json:"created_at"`
}
```

## 3. 功能模块

### 3.1 宝宝系统功能

#### 3.1.1 宝宝降临
- 命令：`宝宝降临`
- 功能：用户创建自己的虚拟宝宝
- 流程：
  1. 检查用户是否已有宝宝
  2. 创建新宝宝记录
  3. 设置初始属性（名字、生日、成长值等）

#### 3.1.2 宝宝学习
- 命令：`宝宝学习`
- 功能：宝宝通过学习获得成长值
- 奖励：每次学习增加100点成长值

#### 3.1.3 宝宝打工
- 命令：`宝宝打工`
- 功能：宝宝通过打工获得成长值和积分
- 条件：宝宝年龄至少1个月
- 奖励：每次打工增加150点成长值和50积分

#### 3.1.4 宝宝互动
- 命令：`宝宝互动`
- 功能：用户与宝宝互动增加亲密度
- 奖励：每次互动增加50点成长值

#### 3.1.5 宝宝商城
- 命令：`宝宝商城`
- 功能：用户可以使用积分购买宝宝用品
- 商品列表：
  - 奶瓶：50积分，增加100成长值
  - 玩具车：100积分，增加200成长值
  - 故事书：150积分，增加300成长值
  - 新衣服：200积分，增加400成长值

#### 3.1.6 宝宝改名
- 命令：`宝宝改名+新名字`
- 功能：修改宝宝的名字
- 条件：名字长度在2-10个字符之间

#### 3.1.7 成长值系统
- 规则：每1000点成长值增加1天年龄
- 等级提升：每30天年龄提升1级
- 自动成长：每日自动增加50点成长值

#### 3.1.8 生日系统
- 规则：每年生日时自动触发生日事件
- 奖励：系统会记录生日事件并给予额外成长值

#### 3.1.9 超管功能
- 命令：`开启宝宝系统`/`关闭宝宝系统`
- 功能：超级管理员可以控制宝宝系统的开关
- 命令：`抛弃宝宝QQ`
- 功能：超级管理员可以处理不当使用宝宝系统的用户

### 3.2 结婚系统功能

#### 3.2.1 婚礼物品购买
- 命令：`购买婚纱`/`购买婚戒`
- 功能：用户购买婚礼所需物品
- 流程：
  1. 检查用户积分
  2. 扣除相应积分
  3. 记录用户购买的物品
- 作用：购买的婚礼物品将用于后续的结婚流程

#### 3.2.2 求婚与结婚
- 命令：`求婚QQ`（QQ为对方的QQ号）
- 功能：向指定用户发送求婚请求
- 流程：
  1. 检查双方是否单身
  2. 创建求婚记录
  3. 等待对方回应

- 命令：`结婚QQ`/`办理结婚证QQ`（QQ为求婚者的QQ号）
- 功能：接受求婚并完成结婚手续
- 流程：
  1. 检查求婚记录
  2. 更新婚姻状态
  3. 创建婚姻记录
  4. 发放结婚证书

#### 3.2.3 离婚
- 命令：`离婚QQ`/`办理离婚证QQ`（QQ为配偶的QQ号）
- 功能：结束当前婚姻关系
- 流程：
  1. 检查用户婚姻状态
  2. 更新婚姻状态为离婚
  3. 记录离婚日期
  4. 发放离婚证

#### 3.2.4 婚姻证件查询
- 命令：`我的结婚证`
- 功能：查询用户的婚姻证件信息
- 返回：婚姻状态、结婚日期、配偶信息、婚姻时长等详细信息

#### 3.2.5 喜糖与红包
- 命令：`发喜糖数量`
- 功能：向群成员发放指定数量的喜糖
- 流程：
  1. 检查用户喜糖数量
  2. 扣除相应喜糖
  3. 群成员可抢食喜糖

- 命令：`吃喜糖`
- 功能：抢食群内的喜糖
- 奖励：随机获得积分或其他奖励

- 命令：`我的喜糖`
- 功能：查询用户当前拥有的喜糖数量

- 命令：`我的红包`
- 功能：查询用户当前拥有的红包数量

#### 3.2.6 另一半互动
- 命令：`另一半签到`
- 功能：为配偶签到并获得奖励
- 奖励：为配偶增加签到积分和成长值

- 命令：`另一半抢楼`
- 功能：为配偶参与抢楼活动并获得奖励
- 规则：与普通抢楼活动类似，但奖励归属于配偶

- 命令：`另一半抢红包`
- 功能：为配偶抢红包并获得奖励
- 规则：自动为配偶参与群内红包抢取

- 命令：`我的对象`
- 功能：召唤或查询当前配偶信息
- 返回：配偶的QQ号、昵称、在线状态等信息

#### 3.2.7 甜蜜爱心系统
- 命令：`我的甜蜜爱心`
- 功能：查询用户当前拥有的甜蜜爱心数量

- 命令：`甜蜜爱心说明`
- 功能：查看甜蜜爱心的获取和使用说明
- 内容：包含甜蜜爱心的来源、用途和使用规则

- 命令：`赠送甜蜜爱心QQ`（QQ为接收方的QQ号）
- 功能：向指定用户赠送甜蜜爱心
- 规则：
  1. 检查用户是否拥有足够的甜蜜爱心
  2. 扣除相应数量的甜蜜爱心
  3. 增加接收方的甜蜜爱心数量

- 命令：`使用甜蜜抽奖`
- 功能：使用甜蜜爱心进行抽奖
- 规则：每次抽奖消耗10个甜蜜爱心
- 奖励：随机获得积分、道具或其他特殊奖励

#### 3.2.8 结婚福利
- 命令：`领取结婚福利`
- 功能：已婚用户可以领取专属福利
- 规则：
  1. 每周可领取一次
  2. 结婚时间越长，福利越丰厚
- 奖励：积分、甜蜜爱心等专属奖励

## 4. 命令列表

### 4.1 宝宝系统命令
| 命令 | 功能描述 | 权限要求 |
|------|----------|----------|
| 宝宝降临 | 创建虚拟宝宝 | 普通用户 |
| 我的宝宝 | 查询宝宝信息 | 普通用户 |
| 宝宝学习 | 宝宝学习获得成长值 | 普通用户 |
| 宝宝打工 | 宝宝打工获得成长值和积分 | 普通用户 |
| 宝宝互动 | 与宝宝互动增加成长值 | 普通用户 |
| 宝宝商城 | 查看可购买的宝宝用品 | 普通用户 |
| 购买+商品编号 | 购买宝宝用品 | 普通用户 |
| 宝宝改名+新名字 | 修改宝宝名字 | 普通用户 |
| 拐卖宝宝说明 | 查看宝宝系统使用规范 | 普通用户 |
| 开启宝宝系统 | 开启宝宝系统 | 超级管理员 |
| 关闭宝宝系统 | 关闭宝宝系统 | 超级管理员 |
| 抛弃宝宝QQ | 处理不当使用宝宝系统的用户（超管专用） | 超级管理员 |

### 4.2 结婚系统命令
| 命令 | 功能描述 | 权限要求 |
|------|----------|----------|
| 购买婚纱 | 购买婚礼婚纱 | 普通用户 |
| 购买婚戒 | 购买婚礼婚戒 | 普通用户 |
| 求婚QQ | 向指定用户发送求婚请求（QQ为对方QQ号） | 普通用户 |
| 结婚QQ | 接受求婚并完成结婚手续（QQ为求婚者QQ号） | 普通用户 |
| 离婚QQ | 结束当前婚姻关系（QQ为配偶QQ号） | 普通用户 |
| 我的结婚证 | 查询婚姻证件信息 | 普通用户 |
| 发喜糖数量 | 向群成员发送指定数量的喜糖 | 普通用户 |
| 吃喜糖 | 抢食群内的喜糖 | 普通用户 |
| 办理结婚证QQ | 办理结婚登记（QQ为求婚者QQ号） | 普通用户 |
| 办理离婚证QQ | 办理离婚登记（QQ为配偶QQ号） | 普通用户 |
| 另一半签到 | 为配偶签到并获得奖励 | 普通用户 |
| 另一半抢楼 | 为配偶抢楼并获得奖励 | 普通用户 |
| 另一半抢红包 | 为配偶抢红包并获得奖励 | 普通用户 |
| 我的对象 | 召唤或查询当前配偶信息 | 普通用户 |
| 我的喜糖 | 查询当前拥有的喜糖数量 | 普通用户 |
| 我的红包 | 查询当前拥有的红包数量 | 普通用户 |
| 我的甜蜜爱心 | 查询当前拥有的甜蜜爱心数量 | 普通用户 |
| 甜蜜爱心说明 | 查看甜蜜爱心的获取和使用说明 | 普通用户 |
| 赠送甜蜜爱心QQ | 向指定用户赠送甜蜜爱心（QQ为接收方QQ号） | 普通用户 |
| 使用甜蜜抽奖 | 使用甜蜜爱心进行抽奖 | 普通用户 |
| 领取结婚福利 | 领取结婚专属福利 | 普通用户 |

## 5. 实现细节

### 5.1 插件架构

宝宝系统和结婚系统均采用BotMatrix的插件架构实现，遵循`plugin.Plugin`接口：

```go
type Plugin interface {
    Name() string
    Description() string
    Version() string
    Init(robot Robot)
}
```

每个系统都有独立的插件文件：
- `src/BotWorker/plugins/baby.go` - 宝宝系统实现
- `src/BotWorker/plugins/marriage.go` - 结婚系统实现

### 5.2 命令解析

系统使用基于正则表达式的命令解析器，通过`CommandParser`结构体实现命令识别和参数提取：

```go
type CommandParser struct {}

func (p *CommandParser) MatchCommand(cmd string, message string) (bool, []string) {
    // 命令匹配逻辑
}

func (p *CommandParser) MatchCommandWithParams(pattern string, message string) (bool, []string) {
    // 带参数的命令匹配逻辑
}
```

### 5.3 数据库集成

两个系统均使用全局数据库连接`GlobalDB`进行数据持久化：

```go
// 在utils.go中定义的全局变量
var GlobalDB *sql.DB
```

系统初始化时会检查数据库连接状态，若未连接则使用模拟数据进行演示。

### 5.4 用户权限管理

系统区分普通用户和超级管理员权限：

```go
func (p *BabyPlugin) isSuperAdmin(userID string) bool {
    // 检查用户是否为超级管理员
}
```

超级管理员可以执行系统开关、用户管理等特殊操作。

## 6. 配置说明

### 6.1 宝宝系统配置

| 配置项 | 类型 | 默认值 | 描述 |
|--------|------|--------|------|
| IsEnabled | bool | true | 宝宝系统是否开启 |
| GrowthRate | int | 1000 | 每增加多少成长值提升1天年龄 |

### 6.2 结婚系统配置

| 配置项 | 类型 | 默认值 | 描述 |
|--------|------|--------|------|
| IsEnabled | bool | true | 结婚系统是否开启 |
| SweetsCost | int | 100 | 发送喜糖的成本 |
| RedPacketCost | int | 200 | 发送红包的成本 |

## 7. 系统扩展

### 7.1 宝宝系统扩展

- 可以添加更多宝宝职业和学习科目
- 可以增加宝宝技能系统
- 可以实现宝宝之间的互动功能

### 7.2 结婚系统扩展

- 可以添加婚礼仪式功能
- 可以实现结婚纪念日系统
- 可以增加夫妻共同任务和挑战

## 8. 测试服功能

测试服功能允许用户体验机器人的新功能，主要特点包括：

- **自由切换**：用户可以通过 `切换测试服` 命令自由开启或关闭测试服功能
- **状态查询**：通过 `测试服状态` 命令查看当前测试服状态
- **详细说明**：通过 `测试服说明` 命令查看测试服功能的完整说明
- **新功能体验**：在测试服环境中提前体验机器人的新功能

更多详细信息请参考 [测试服功能文档](src/BotWorker/docs/test_server.md)

## 9. 注意事项

1. 系统依赖全局数据库连接，确保主程序正确初始化数据库
2. 命令参数格式需严格遵循要求
3. 超级管理员权限需谨慎分配
4. 系统使用模拟数据时，所有操作不会实际修改数据库

## 10. 更新日志

### 版本 2.0.0
- 完善宝宝系统功能，包括成长值和生日系统
- 完善结婚系统功能，添加甜蜜爱心和结婚福利
- 实现68个小游戏功能，优化菜单系统
- 添加测试服功能，允许用户体验新功能
- 优化命令解析和处理逻辑
- 增强系统稳定性和性能

### 版本 1.0.0
- 实现宝宝系统核心功能
- 实现结婚系统核心功能
- 支持基本的命令解析和处理
- 提供模拟数据用于演示