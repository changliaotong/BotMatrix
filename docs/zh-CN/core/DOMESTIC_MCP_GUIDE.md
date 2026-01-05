# 国内 MCP 生态适配与推荐指南 (Domestic MCP Guide)

针对国外 MCP 服务（如 Google Search, Slack, Brave Search 等）在境内访问受限或数据不合规的问题，BotMatrix 推荐以下本土化适配方案与 MCP 服务。

## 1. 核心推荐：国内可用 MCP 服务

### 1.1 搜索与信息获取 (Search)
- **Bing Search MCP**: 微软 Bing 在国内可用性较高，推荐使用基于 Bing API 的 MCP Server。
- **Serper / Bocha (博查)**: 国内领先的 AI 搜索接口，可通过我们的 `GenericWebhookMCPHost` 快速接入。
- **知乎/公众号 爬虫 MCP**: 社区中已有针对知乎、微信公众号内容的开源 MCP 适配器，用于获取中文深度内容。

### 1.2 办公与协作 (Office & Collaboration)
- **飞书 (Feishu/Lark) MCP**: 
    - **功能**: 读取多维表格、发送群消息、管理日历。
    - **推荐理由**: 接口友好，是国内最适合做 MCP 化的办公软件。
- **钉钉 (DingTalk) MCP**: 
    - **功能**: 机器人通知、文档读写。
- **企业微信 MCP**: 针对私域流量管理的工具集。

### 1.3 基础工具 (Infrastructure)
- **本地文件系统 (Local Filesystem)**: 无需网络，直接操作本地文档、代码库。
- **数据库 (MySQL/PostgreSQL)**: 接入公司内部数据库，实现自然语言查询 SQL。
- **Python 执行器 (Code Interpreter)**: 在本地安全沙箱执行 Python 脚本，进行数据处理。

---

## 2. BotMatrix 的本土化增强方案

为了最大化发挥 MCP 在国内环境的价值，BotMatrix 提供了以下特有功能：

### 2.1 聚合搜索桥接 (Unified Search Bridge)
我们建议用户建立一个“搜索聚合 MCP”，将百度、博查、微信搜索整合在一个工具下。AI 会根据问题自动选择最合适的国内搜索源。

### 2.2 跨内网穿透 (Intranet Piercing)
针对部署在公司内网的业务系统（如 ERP、CRM），BotMatrix 的 **BotWorker** 可以作为 MCP Server 的宿主，通过加密隧道将内网能力安全地提供给云端大模型调用。

### 2.3 敏感词过滤 (Domestic Compliance)
在调用任何 MCP 工具前，BotMatrix 会强制经过 [privacy.go](file:///D:/projects/BotMatrix/src/Common/ai/privacy.go) 和合规性审查模块，确保返回的搜索内容符合国内监管要求。

---

## 3. 开发者建议：如何编写国内 MCP？

如果你需要接入一个国内特有的业务系统（如金蝶 ERP），推荐以下路径：
1.  **极简模式**: 使用 BotMatrix 提供的 `GenericWebhookMCPHost`，只需提供 API 地址和 Token 即可接入。
2.  **插件模式**: 在 BotMatrix 的 Skill Center 编写 Python/Go 插件，系统会自动将其暴露为标准 MCP 接口。
3.  **社区模式**: 关注 GitHub 上的 `mcp-servers` 组织，寻找国内开发者维护的适配器。

---
*BotMatrix 社区委员会 - 2026-01-03*
