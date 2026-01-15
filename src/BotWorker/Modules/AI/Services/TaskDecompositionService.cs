using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Modules.AI.Interfaces;
using Newtonsoft.Json;
using Microsoft.Extensions.Logging;

namespace BotWorker.Modules.AI.Services
{
    public class TaskDecompositionService : ITaskDecompositionService
    {
        private readonly IAIService _aiService;
        private readonly ILogger<TaskDecompositionService> _logger;

        public TaskDecompositionService(IAIService aiService, ILogger<TaskDecompositionService> logger)
        {
            _aiService = aiService;
            _logger = logger;
        }

        public async Task<List<SubTaskInfo>> DecomposeAsync(string complexTask)
        {
            var prompt = $@"你是一个高级项目经理和任务架构师。你的目标是将用户提出的复杂任务拆分为一系列可由不同角色执行的子任务。

## 用户任务
{complexTask}

## 拆分规则
1. 识别出任务中的逻辑步骤。
2. 为每个步骤指定一个最合适的角色 (JobKey)。可选角色目前主要有：
   - 'dev_orchestrator': 负责代码编写、项目重构、Git 操作、自动化构建。
   - 'code_reviewer': 负责代码审查和质量控制。
   - 'task_planner': 负责宏观规划（如果任务非常大）。
3. 明确子任务之间的依赖关系（如果有）。
4. 每个子任务的描述应该是自包含的，能够让执行者明白上下文。

## 输出格式
请直接输出 JSON 数组，格式如下：
[
  {{
    ""Title"": ""子任务标题"",
    ""Description"": ""详细描述"",
    ""JobKey"": ""角色标识"",
    ""Dependencies"": [""依赖的任务标题1""]
  }}
]

注意：只输出 JSON 内容，不要包含任何 Markdown 代码块标签或其他解释文字。";

            try
            {
                var response = await _aiService.ChatAsync(prompt);
                // 清理可能的 Markdown 格式
                response = response.Trim();
                if (response.StartsWith("```json")) response = response.Substring(7);
                if (response.StartsWith("```")) response = response.Substring(3);
                if (response.EndsWith("```")) response = response.Substring(0, response.Length - 3);
                response = response.Trim();

                var subTasks = JsonConvert.DeserializeObject<List<SubTaskInfo>>(response);
                return subTasks ?? new List<SubTaskInfo>();
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Failed to decompose task: {Task}", complexTask);
                // 如果拆分失败，退化为单一任务
                return new List<SubTaskInfo>
                {
                    new SubTaskInfo
                    {
                        Title = "执行原始任务",
                        Description = complexTask,
                        JobKey = "dev_orchestrator"
                    }
                };
            }
        }
    }
}
