# AI 系统 Token 计量与记录机制

本文档详细说明了 BotMatrix 系统中 Token 的计算、计量及记录机制，明确了“估算数据”与“真实数据”的区别。

## 1. 核心计量机制：API 真实回传

系统在统计 Token 消耗（用于计费和审计）时，**完全依赖大模型 API 服务商回传的真实数据**。

### 实现原理
在 [openai_adapter.go](file:///D:/projects/BotMatrix/src/Common/ai/openai_adapter.go) 中，系统直接解析 API 响应体中的 `usage` 字段：

```go
// 原始 API 响应解析
var result ChatResponse
if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
    return nil, err
}
// 此时 result.Usage 包含的是 API 服务商计算的精确 Token 数
```

### 记录位置
真实 Token 数据会通过以下两个途径展示和记录：
1. **系统日志**: 在 [ai_service.go](file:///D:/projects/BotMatrix/src/BotNexus/internal/app/ai_service.go#L259) 中打印。
2. **数据库持久化**: 异步写入 `ai_usage_logs` 表，包含 `input_tokens`, `output_tokens` 和 `total_tokens`。

---

## 2. 预处理机制：本地估算 (ContextManager)

在请求发送**之前**，系统会进行一次本地 Token 估算。

### 目的
- **防止溢出**: 在发送请求前判断当前对话历史是否超过了模型窗口上限。
- **自动修剪**: 如果估算值过高，[ContextManager](file:///D:/projects/BotMatrix/src/Common/ai/context_manager.go) 会自动删减旧消息。

### 计算公式
系统采用了一种保守且高效的估算算法：
> **Token 估算值 = (字符总长度 / 4) + 基础开销**

### 区别说明
| 特性 | 本地估算 (Estimate) | API 真实回传 (Usage) |
| :--- | :--- | :--- |
| **发生阶段** | 请求发送前 | 响应接收后 |
| **计算依据** | 字符长度简单换算 | 模型专用 Tokenizer 精确计算 |
| **准确度** | 近似值 (通常偏大以保证安全) | 100% 准确 (权威数据) |
| **主要用途** | 上下文剪裁、请求前校验 | 计费、成本统计、使用记录 |

---

## 3. 常见问题 (FAQ)

### Q: 为什么不使用本地 Tokenizer 进行精确计算？
A: 不同模型的 Tokenizer 实现不同（如 GPT-4 与 Llama-3 不同）。为了保持系统的轻量化和通用性，我们在请求前使用估算值来保证安全，在请求后使用 API 返回的真实值进行精确统计。

### Q: 自主 Agent 循环 (ChatAgent) 如何统计 Token？
A: `ChatAgent` 会进行多次 `Chat` 调用。每一次调用的真实 Token 都会被单独记录到 `ai_usage_logs` 中，因此你可以通过 `session_id` 聚合出整个任务的总消耗。
