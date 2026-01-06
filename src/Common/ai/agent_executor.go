package ai

import (
	clog "BotMatrix/common/log"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"go.uber.org/zap"
)

// AgentExecutor 负责执行 Agent 的推理循环
// 它实现了 ReAct (Reasoning + Acting) 模式，并集成了 Sandbox 等工具的调用
type AgentExecutor struct {
	aiService AIService
	modelID   uint
	botID     string
	userID    string
	sessionID string
}

func NewAgentExecutor(aiService AIService, modelID uint, botID, userID, sessionID string) *AgentExecutor {
	return &AgentExecutor{
		aiService: aiService,
		modelID:   modelID,
		botID:     botID,
		userID:    userID,
		sessionID: sessionID,
	}
}

// Execute 启动 Agent 循环
func (e *AgentExecutor) Execute(ctx context.Context, initialMessages []Message, tools []Tool) (*ChatResponse, error) {
	const maxIterations = 15 // 增加迭代次数以支持复杂任务
	currentMessages := initialMessages

	var finalResp *ChatResponse

	// 注入系统级提示词，引导模型进行 ReAct 思考
	// 注意：如果模型本身微调过，可能不需要这么详细的 prompt
	systemPrompt := `你是一个具有自主规划和执行能力的 AI 智能体。
当面对复杂任务时，请遵循以下步骤：
1. 思考 (Thought): 分析当前状态，决定下一步做什么。
2. 规划 (Plan): 如果需要，列出后续步骤。
3. 行动 (Action): 调用合适的工具（如 sandbox_exec 执行代码）。
4. 观察 (Observation): 查看工具的返回结果。
5. 反思 (Reflection): 根据结果修正计划或得出结论。

你可以使用 Sandbox 来执行代码、测试脚本、处理文件。
请务必确保代码在 Sandbox 中运行，不要试图在宿主机运行。
`
	// 检查是否已经有了 System Prompt
	hasSystem := false
	for _, m := range currentMessages {
		if m.Role == RoleSystem {
			hasSystem = true
			break
		}
	}
	if !hasSystem {
		currentMessages = append([]Message{{Role: RoleSystem, Content: systemPrompt}}, currentMessages...)
	}

	for i := 0; i < maxIterations; i++ {
		clog.Info("[AgentExecutor] Iteration", zap.Int("step", i+1), zap.String("session", e.sessionID))

		// 上下文传递
		stepCtx := context.WithValue(ctx, "sessionID", e.sessionID)
		stepCtx = context.WithValue(stepCtx, "step", i)
		stepCtx = context.WithValue(stepCtx, "botID", e.botID)

		// TODO: 传递 userIDNum, orgIDNum 等，如果需要鉴权

		// 1. 调用 LLM
		resp, err := e.aiService.Chat(stepCtx, e.modelID, currentMessages, tools)
		if err != nil {
			return nil, fmt.Errorf("LLM chat failed: %v", err)
		}
		finalResp = resp

		if len(resp.Choices) == 0 {
			break
		}
		choice := resp.Choices[0]

		// 2. 检查是否结束
		// 如果模型认为已经完成 (Stop)，且没有 ToolCalls，则结束
		if choice.FinishReason != "tool_calls" && len(choice.Message.ToolCalls) == 0 {
			// 有时候模型虽然输出了内容，但可能还在思考中，这里简单判断
			// 如果内容包含 "Final Answer" 或者看起来像结束语，则退出
			// 但对于通用模型，通常 finish_reason=stop 就是结束
			break
		}

		// 3. 记录助手消息
		currentMessages = append(currentMessages, choice.Message)

		// 4. 执行工具
		if len(choice.Message.ToolCalls) > 0 {
			toolResults := e.executeToolsParallel(stepCtx, choice.Message.ToolCalls)

			// 5. 追加工具结果
			for _, res := range toolResults {
				currentMessages = append(currentMessages, Message{
					Role:       RoleTool,
					Content:    res.Output,
					ToolCallID: res.ToolCallID,
					Name:       res.Name,
				})
			}
		}
	}

	return finalResp, nil
}

type toolExecutionResult struct {
	ToolCallID string
	Name       string
	Output     string
	Error      error
}

func (e *AgentExecutor) executeToolsParallel(ctx context.Context, toolCalls []ToolCall) []toolExecutionResult {
	var wg sync.WaitGroup
	results := make([]toolExecutionResult, len(toolCalls))

	for i, tc := range toolCalls {
		wg.Add(1)
		go func(index int, t ToolCall) {
			defer wg.Done()

			clog.Info("[AgentExecutor] Executing tool", zap.String("name", t.Function.Name))

			// 构造结果
			res := toolExecutionResult{
				ToolCallID: t.ID,
				Name:       t.Function.Name,
			}

			// 解析参数
			var args map[string]any
			if err := json.Unmarshal([]byte(t.Function.Arguments), &args); err != nil {
				res.Error = err
				res.Output = fmt.Sprintf("Error parsing arguments: %v", err)
				results[index] = res
				return
			}

			// 调用 MCP
			mcpMgr := e.aiService.GetMCPManager()
			if mcpMgr == nil {
				res.Error = fmt.Errorf("MCP Manager not available")
				res.Output = "Error: MCP Manager not available"
				results[index] = res
				return
			}

			toolResult, err := mcpMgr.CallTool(ctx, t.Function.Name, args)
			if err != nil {
				res.Error = err
				res.Output = fmt.Sprintf("Error executing tool: %v", err)
			} else {
				// 序列化结果
				// MCP 返回的通常是 struct，需要转 json string
				// 或者是 MCPCallToolResponse
				if mcpResp, ok := toolResult.(MCPCallToolResponse); ok {
					// 提取 Content 中的 Text
					var sb strings.Builder
					for _, c := range mcpResp.Content {
						if c.Type == "text" {
							sb.WriteString(c.Text)
							sb.WriteString("\n")
						}
					}
					res.Output = strings.TrimSpace(sb.String())
				} else if mcpRespPtr, ok := toolResult.(*MCPCallToolResponse); ok {
					var sb strings.Builder
					for _, c := range mcpRespPtr.Content {
						if c.Type == "text" {
							sb.WriteString(c.Text)
							sb.WriteString("\n")
						}
					}
					res.Output = strings.TrimSpace(sb.String())
				} else {
					b, _ := json.Marshal(toolResult)
					res.Output = string(b)
				}
			}
			results[index] = res
		}(i, tc)
	}

	wg.Wait()
	return results
}
