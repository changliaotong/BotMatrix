using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Models.Evolution;
using Microsoft.Extensions.Logging;
using System.Text.Json;
using System.Text;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;

namespace BotWorker.Modules.AI.Services
{
    public class EvolutionService : IEvolutionService
    {
        private readonly IAIService _aiService;
        private readonly IJobService _jobService;
        private readonly IJobDefinitionRepository _jobRepository;
        private readonly ITaskRecordRepository _taskRepository;
        private readonly ITaskStepRepository _stepRepository;
        private readonly IEmployeeInstanceRepository _employeeRepository;
        private readonly ILogger<EvolutionService> _logger;

        public EvolutionService(
            IAIService aiService, 
            IJobService jobService, 
            IJobDefinitionRepository jobRepository,
            ITaskRecordRepository taskRepository,
            ITaskStepRepository stepRepository,
            IEmployeeInstanceRepository employeeRepository,
            ILogger<EvolutionService> logger)
        {
            _aiService = aiService;
            _jobService = jobService;
            _jobRepository = jobRepository;
            _taskRepository = taskRepository;
            _stepRepository = stepRepository;
            _employeeRepository = employeeRepository;
            _logger = logger;
        }

        public async Task<bool> EvolveJobAsync(string jobId)
        {
            var job = await _jobService.GetJobAsync(jobId);
            if (job == null) return false;

            // 1. 获取该岗位下的所有员工
            var employees = await _employeeRepository.GetByJobIdAsync(job.Id);
            var employeeIds = employees.Select(e => e.Id).ToList();

            if (!employeeIds.Any()) return false;

            // 2. 获取这些员工最近未被进化的已完成任务
            // 注意：这里需要 Repository 支持按员工 ID 列表查询，或者循环查询
            var recentTasks = new List<TaskRecord>();
            foreach (var empId in employeeIds)
            {
                var tasks = await _taskRepository.GetByAssigneeIdAsync(empId);
                recentTasks.AddRange(tasks.Where(t => t.Status == "completed").Take(5));
            }

            if (recentTasks.Count < 3) 
            {
                _logger.LogInformation("[EvolutionService] Not enough tasks for evolution of job {JobId}", jobId);
                return false;
            }

            // 3. 收集执行反馈
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
                var steps = await _stepRepository.GetByTaskIdAsync(task.Id);
                foreach (var step in steps)
                {
                    if (!string.IsNullOrEmpty(step.OutputData))
                    {
                        sb.AppendLine($"- [步骤: {step.Name}] 状态: {step.Status}");
                        if (step.Status == "failed")
                        {
                            sb.AppendLine($"  - 错误内容: {step.ErrorMessage}");
                        }
                    }
                }
            }

            // 4. 调用 LLM 进行进化分析
            var evolutionPrompt = $@"你现在是数字员工进化引擎 (Evolution Engine)。
请分析以下岗位的执行数据，并提出优化建议。

{sb}

## 进化任务
请判断当前的岗位约束 (Constraints) 或工作流 (Workflow) 是否需要优化？
如果需要，请提供更新后的 JSON。

## 输出格式 (必须是 JSON)
{{
  ""needs_update"": true,
  ""reason"": ""为什么要更新"",
  ""new_constraints"": ""优化后的约束条件"",
  ""new_workflow"": ""优化后的工作流 JSON 字符串""
}}";

            var response = await _aiService.ChatAsync(evolutionPrompt, job.ModelSelectionStrategy);
            
            // 5. 解析并应用进化
            var jsonStart = response.IndexOf("{");
            var jsonEnd = response.LastIndexOf("}");
            if (jsonStart >= 0 && jsonEnd > jsonStart)
            {
                var jsonStr = response.Substring(jsonStart, jsonEnd - jsonStart + 1);
                try
                {
                    var result = JsonSerializer.Deserialize<EvolutionProposal>(jsonStr, new JsonSerializerOptions { PropertyNameCaseInsensitive = true });
                    
                    if (result != null && result.Needs_Update)
                    {
                        job.Constraints = result.New_Constraints;
                        job.Workflow = result.New_Workflow;
                        
                        // 允许进化建议修改模型选择策略
                        if (!string.IsNullOrEmpty(result.New_Model_Strategy))
                        {
                            job.ModelSelectionStrategy = result.New_Model_Strategy;
                        }

                        job.Version++;
                        await _jobRepository.UpdateAsync(job);

                        _logger.LogInformation("[EvolutionService] Job {JobId} evolved to version {Version}. Strategy: {Strategy}. Reason: {Reason}", 
                            jobId, job.Version, job.ModelSelectionStrategy, result.Reason);
                    }
                    return true;
                }
                catch (Exception ex)
                {
                    _logger.LogError(ex, "[EvolutionService] Error parsing evolution proposal: {Response}", response);
                }
            }

            return false;
        }

        public async Task EvolveAllJobsAsync()
        {
            var jobs = await _jobRepository.GetActiveJobsAsync();
            foreach (var job in jobs)
            {
                // 使用 JobKey 进行进化
                await EvolveJobAsync(job.JobKey);
            }
        }

        private class EvolutionProposal
        {
            public bool Needs_Update { get; set; }
            public string Reason { get; set; } = string.Empty;
            public string New_Constraints { get; set; } = string.Empty;
            public string New_Workflow { get; set; } = string.Empty;
            public string New_Model_Strategy { get; set; } = string.Empty;
        }
    }
}
