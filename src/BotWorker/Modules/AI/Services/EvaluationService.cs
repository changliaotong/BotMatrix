using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Models.Evolution;
using Microsoft.Extensions.Logging;
using System.Text.Json;
using System;
using System.Threading.Tasks;

namespace BotWorker.Modules.AI.Services
{
    public class EvaluationService : IEvaluationService
    {
        private readonly IAIService _aiService;
        private readonly ITaskRecordRepository _taskRepository;
        private readonly ITaskStepRepository _stepRepository;
        private readonly ILogger<EvaluationService> _logger;

        public EvaluationService(
            IAIService aiService, 
            ITaskRecordRepository taskRepository,
            ITaskStepRepository stepRepository,
            ILogger<EvaluationService> logger)
        {
            _aiService = aiService;
            _taskRepository = taskRepository;
            _stepRepository = stepRepository;
            _logger = logger;
        }

        public async Task<bool> EvaluateStepAsync(TaskStep step, string taskPrompt)
        {
            try
            {
                var evalPrompt = $@"你现在是一名资深质量审计专家 (Reviewer)。
请评估以下数字员工的执行结果。

## 原始任务需求
{taskPrompt}

## 执行步骤
{step.Name}

## LLM 输出内容
{step.OutputData}

## 评价标准
1. 是否完成了用户要求的目标？
2. 是否遵循了岗位约束和工作流？
3. 输出是否专业、准确、无误？

## 输出格式 (必须是 JSON)
{{
  ""score"": 0-100 之间的整数,
  ""feedback"": ""简短的评估建议"",
  ""is_success"": true/false
}}";

                var response = await _aiService.ChatAsync(evalPrompt);
                
                // 尝试解析 JSON
                var jsonStart = response.IndexOf("{");
                var jsonEnd = response.LastIndexOf("}");
                if (jsonStart >= 0 && jsonEnd > jsonStart)
                {
                    var jsonStr = response.Substring(jsonStart, jsonEnd - jsonStart + 1);
                    var result = JsonSerializer.Deserialize<EvaluationResult>(jsonStr, new JsonSerializerOptions { PropertyNameCaseInsensitive = true });
                    
                    if (result != null)
                    {
                        // 映射评估结果到 TaskStep (原本 TaskExecution 的字段在 TaskStep 中可能需要对应)
                        step.Status = result.Is_Success ? "completed" : "failed";
                        if (!result.Is_Success)
                        {
                            step.ErrorMessage = result.Feedback;
                        }
                        await _stepRepository.UpdateAsync(step);
                        return true;
                    }
                }
                
                _logger.LogWarning("[EvaluationService] Failed to parse evaluation response for Step {Id}", step.Id);
                return false;
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "[EvaluationService] Error evaluating step {Id}", step.Id);
                return false;
            }
        }

        public async Task<bool> EvaluateTaskResultAsync(TaskRecord task)
        {
            try
            {
                task.Status = "completed";
                return await _taskRepository.UpdateAsync(task);
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "[EvaluationService] Error completing task {Id}", task.Id);
                return false;
            }
        }

        private class EvaluationResult
        {
            public int Score { get; set; }
            public string Feedback { get; set; } = string.Empty;
            public bool Is_Success { get; set; }
        }
    }
}
