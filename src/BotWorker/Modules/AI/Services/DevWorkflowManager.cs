using BotWorker.Modules.AI.Interfaces;
using Microsoft.Extensions.Logging;
using System.Text.Json;

namespace BotWorker.Modules.AI.Services
{
    public class DevWorkflowManager : IDevWorkflowManager
    {
        private readonly IAgentExecutor _agentExecutor;
        private readonly ILogger<DevWorkflowManager> _logger;

        public DevWorkflowManager(IAgentExecutor agentExecutor, ILogger<DevWorkflowManager> logger)
        {
            _agentExecutor = agentExecutor;
            _logger = logger;
        }

        public async Task<bool> StartDevProjectAsync(string requirementDoc, string projectPath)
        {
            _logger.LogInformation("[DevWorkflow] Starting new project in {Path}", projectPath);

            try
            {
                // 1. RA: 需求分析
                _logger.LogInformation("[DevWorkflow] Phase 1: Requirements Analysis...");
                var specJson = await _agentExecutor.ExecuteByJobAsync("dev_ra", $"请分析以下文档并输出技术规格 JSON：\n{requirementDoc}");
                
                // 2. SA: 架构设计
                _logger.LogInformation("[DevWorkflow] Phase 2: Architecture Design...");
                var structureJson = await _agentExecutor.ExecuteByJobAsync("dev_architect", $"根据规格说明书设计项目结构：\n{specJson}");
                
                // 3. 解析任务列表 (简化处理：假设架构师输出了待编写的文件列表)
                // 在实际复杂场景中，这里需要更强的 JSON 解析和循环处理逻辑
                var tasks = ParseTasks(structureJson);

                // 4. SD: 并行/顺序编码
                foreach (var task in tasks)
                {
                    _logger.LogInformation("[DevWorkflow] Phase 3: Coding file {File}...", task.FileName);
                    var code = await _agentExecutor.ExecuteByJobAsync("dev_coder", $"请编写文件 {task.FileName}，功能描述：{task.Description}\n技术规格：{specJson}");
                    
                    // 写入物理文件
                    var fullPath = Path.Combine(projectPath, task.FileName);
                    Directory.CreateDirectory(Path.GetDirectoryName(fullPath)!);
                    await File.WriteAllTextAsync(fullPath, code);
                }

                _logger.LogInformation("[DevWorkflow] Project completed successfully!");
                return true;
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "[DevWorkflow] Error during auto-dev process");
                return false;
            }
        }

        private List<DevTask> ParseTasks(string structureJson)
        {
            try
            {
                // 尝试解析标准的 JSON 格式
                var jsonStart = structureJson.IndexOf("{");
                var jsonEnd = structureJson.LastIndexOf("}");
                if (jsonStart >= 0 && jsonEnd > jsonStart)
                {
                    var cleanJson = structureJson.Substring(jsonStart, jsonEnd - jsonStart + 1);
                    var doc = JsonDocument.Parse(cleanJson);
                    if (doc.RootElement.TryGetProperty("tasks", out var tasksProp))
                    {
                        return JsonSerializer.Deserialize<List<DevTask>>(tasksProp.GetRawText(), new JsonSerializerOptions { PropertyNameCaseInsensitive = true }) ?? new List<DevTask>();
                    }
                }
            }
            catch
            {
                _logger.LogWarning("[DevWorkflow] Failed to parse structure JSON, using fallback tasks.");
            }

            // 简单模拟解析，实际应使用正则或 JSON 反序列化
            return new List<DevTask>
            {
                new DevTask { FileName = "README.md", Description = "项目概览" },
                new DevTask { FileName = "main.py", Description = "主程序业务逻辑实现" },
                new DevTask { FileName = "requirements.txt", Description = "项目依赖列表" }
            };
        }

        private class DevTask
        {
            public string FileName { get; set; } = string.Empty;
            public string Description { get; set; } = string.Empty;
        }
    }
}
