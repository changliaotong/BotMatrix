# 用户自定义 AI 密钥与算力租赁方案规划

## 1. 核心目标
- **用户赋能**：允许用户提交自己的 AI API Key（如 DeepSeek, OpenAI），在聊天时优先使用自己的额度。
- **算力共享**：用户可以将自己用不完的 Key 租赁给系统，供其他没有 Key 的用户使用。
- **激励机制**：当用户的 Key 被租赁使用时，系统给予该用户一定的算力奖励（积分/Token）。

## 2. 功能规划

### 2.1 用户 API Key 管理
- **设置 Key**：
  - 指令：`设置Key [提供商] [ApiKey] [BaseUrl(可选)]`
  - 逻辑：将配置存储在 `UserAIConfig` 表中。
- **查询配置**：
  - 指令：`我的Key`
  - 逻辑：列出用户已配置的所有提供商、Key 的掩码处理、租赁状态及使用统计。

### 2.2 算力租赁机制
- **开启租赁**：
  - 指令：`开启租赁 [提供商]`
  - 逻辑：将 `UserAIConfig.IsLeased` 设置为 `true`。
- **关闭租赁**：
  - 指令：`关闭租赁 [提供商]`
  - 逻辑：将 `UserAIConfig.IsLeased` 设置为 `false`。
- **调度逻辑**：
  - 优先级：用户自有 Key > 租赁池随机 Key > 系统默认 Key。

### 2.3 激励与统计
- **奖励逻辑**：每当租赁池中的某个 Key 被成功调用一次，通过 `UserInfo.AddTokensAsync` 为提供者增加奖励。
- **统计字段**：
  - `UseCount`: 累计使用次数。
  - `LastUsedAt`: 最后一次使用时间。

## 3. 技术实现方案

### 3.1 数据模型 (UserAIConfig.cs)
- 实体类已创建，包含以下核心字段：
  - `UserId`: 用户 ID。
  - `ProviderName`: AI 提供商名称（如 DeepSeek, OpenAI）。
  - `ApiKey`: 加密存储（当前为明文，建议后续加密）。
  - `BaseUrl`: 接口地址。
  - `IsLeased`: 是否参与租赁。
  - `UseCount`: 使用次数统计。

### 3.2 逻辑集成
- **指令集成**：在 `BuiltinCommandMiddleware.cs` 或 `AiConfigMessage.cs` 中实现指令解析。
- **服务集成**：在 `AIService.cs` 的 `ChatWithContextAsync` 中实现多级 Key 查找逻辑。

## 4. 后续扩展
- **余额监控**：自动检测租赁池中的 Key 是否失效或余额不足，并自动禁用。
- **多模型支持**：支持用户为同一提供商配置多个不同模型的特定 Key。
- **加密存储**：对数据库中的 `ApiKey` 进行加密处理，确保用户隐私。
