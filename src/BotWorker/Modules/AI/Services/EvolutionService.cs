using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Models.Evolution;
using Microsoft.Extensions.Logging;
using System.Text.Json;
using System.Text;

namespace BotWorker.Modules.AI.Services
{
    public class EvolutionService : IEvolutionService
    {
        private readonly IAIService _aiService;
        private readonly IJobService _jobService;
        private readonly ILogger<EvolutionService> _logger;

        public EvolutionService(IAIService aiService, IJobService jobService, ILogger<EvolutionService> logger)
        {
            _aiService = aiService;
            _jobService = jobService;
            _logger = logger;
        }

        public async Task<bool> EvolveJobAsync(string jobId)
        {
            var job = await _jobService.GetJobAsync(jobId);
            if (job == null) return false;

            // 1. 获取最近未被进化的任务记录和执行详情
            var recentTasks = await TaskRecord.QueryListAsync<TaskRecord>($@"
                SELECT TOP 10 * FROM {TaskRecord.FullName} 
                WHERE EmployeeId IN (SELECT EmployeeId FROM {EmployeeInstance.FullName} WHERE JobId = '{jobId}')
                AND IsEvolved = 0 AND Status = 'Completed'
                ORDER BY CreatedAt DESC");

            if (recentTasks.Count < 3) 
            {
                _logger.LogInformation("[EvolutionService] Not enough tasks for evolution of job {JobId}", jobId);
                return false;
            }

            // 2. 收集执行反馈和失败案例
            var sb = new StringBuilder();
            sb.AppendLine($"# 岗位当前定义");
            sb.AppendLine($"名称: {job.Name}");
            sb.AppendLine($"目标: {job.Purpose}");
            sb.AppendLine($"约束: {job.Constraints}");
            sb.AppendLine($"工作流: {job.Workflow}");
            sb.AppendLine();
            sb.AppendLine("# 最近执行反馈摘要");

            foreach (var task in recentTasks)
            {
                var executions = await TaskExecution.QueryListAsync<TaskExecution>($"SELECT * FROM {TaskExecution.FullName} WHERE TaskId = '{task.TaskId}'");
                foreach (var exec in executions)
                {
                    sb.AppendLine($"- [得分: {exec.EvaluationScore}] 反馈: {exec.EvaluationFeedback}");
                    if (exec.EvaluationScore < 60)
                    {
                        sb.AppendLine($"  - 失败输入: {exec.InputData}");
                        sb.AppendLine($"  - 错误内容: {exec.ErrorMessage}");
                    }
                }
            }

            // 3. 调用 LLM 进行进化分析
            var evolutionPrompt = $@"你现在是数字员工进化引擎 (Evolution Engine)。
请分析以下岗位的执行数据，并提出优化建议。

{sb}

## 进化任务
请判断当前的岗位约束 (Constraints) 或工作流 (Workflow) 是否需要优化？
如果需要，请提供更新后的 JSON。

## 输出格式 (必须是 JSON)
{{
  ""needs_update"": true/false,
  ""reason"": ""为什么要更新"",
  ""new_constraints"": ""优化后的约束条件"",
  ""new_workflow"": ""优化后的工作流 JSON 字符串""
}}";

            var response = await _aiService.ChatAsync(evolutionPrompt);
            
            // 4. 解析并应用进化
            var jsonStart = response.IndexOf("{");
            var jsonEnd = response.LastIndexOf("}");
            if (jsonStart >= 0 && jsonEnd > jsonStart)
            {
                var jsonStr = response.Substring(jsonStart, jsonEnd - jsonStart + 1);
                var result = JsonSerializer.Deserialize<EvolutionProposal>(jsonStr, new JsonSerializerOptions { PropertyNameCaseInsensitive = true });
                
                if (result != null && result.Needs_Update)
                {
                    job.Constraints = result.New_Constraints;
                    job.Workflow = result.New_Workflow;
                    job.Version++;
                    await job.SaveAsync();

                    _logger.LogInformation("[EvolutionService] Job {JobId} evolved to version {Version}. Reason: {Reason}", jobId, job.Version, result.Reason);
                }

                // 标记任务为已进化
                foreach (var task in recentTasks)
                {
                    task.IsEvolved = true;
                    await task.SaveAsync();
                }
                return true;
            }

            return false;
        }

        public async Task EvolveAllJobsAsync()
        {
            var jobs = await JobDefinition.QueryListAsync<JobDefinition>($"SELECT * FROM {JobDefinition.FullName} WHERE IsActive = 1");
            foreach (var job in jobs)
            {
                await EvolveJobAsync(job.JobId);
            }
        }

        private class EvolutionProposal
        {
            public bool Needs_Update { get; set; }
            public string Reason { get; set; } = string.Empty;
            public string New_Constraints { get; set; } = string.Empty;
            public string New_Workflow { get; set; } = string.Empty;
        }
    }
}
