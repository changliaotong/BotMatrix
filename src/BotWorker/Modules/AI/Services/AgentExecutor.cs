using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Models.Evolution;
using BotWorker.Modules.AI.Models;
using Microsoft.Extensions.Logging;
using System;
using System.Threading.Tasks;

namespace BotWorker.Modules.AI.Services
{
    public interface IAgentExecutor
    {
        Task<string> ExecuteByAgentAsync(Agent agent, string prompt);
        Task<string> ExecuteByAgentGuidAsync(Guid guid, string prompt);
        Task<string> ExecuteByJobAsync(string jobId, string prompt);
        Task<string> ExecuteJobTaskAsync(string jobId, string prompt, IPluginContext context, string? employeeId = null);
    }

    public class AgentExecutor : IAgentExecutor
    {
        private readonly IAIService _aiService;
        private readonly IJobService _jobService;
        private readonly IEmployeeService _employeeService;
        private readonly IEvaluationService _evaluationService;
        private readonly ITaskRecordRepository _taskRepository;
        private readonly ITaskStepRepository _stepRepository;
        private readonly ILogger<AgentExecutor> _logger;

        public AgentExecutor(
            IAIService aiService, 
            IJobService jobService, 
            IEmployeeService employeeService,
            IEvaluationService evaluationService,
            ITaskRecordRepository taskRepository,
            ITaskStepRepository stepRepository,
            ILogger<AgentExecutor> _logger)
        {
            this._aiService = aiService;
            this._jobService = jobService;
            this._employeeService = employeeService;
            this._evaluationService = evaluationService;
            this._taskRepository = taskRepository;
            this._stepRepository = stepRepository;
            this._logger = _logger;
        }

        public async Task<string> ExecuteByAgentAsync(Agent agent, string prompt)
        {
            var systemPrompt = agent.SystemPrompt ?? "You are a helpful assistant.";
            return await _aiService.ChatAsync($"{systemPrompt}\n\nUser: {prompt}");
        }

        public async Task<string> ExecuteByAgentGuidAsync(Guid guid, string prompt)
        {
            // TODO: 从 AgentRepository 获取
            return await _aiService.ChatAsync(prompt);
        }

        public async Task<string> ExecuteByJobAsync(string jobId, string prompt)
        {
            var job = await _jobService.GetJobAsync(jobId);
            if (job == null) return $"Error: Job {jobId} not found.";

            var systemPrompt = $@"你现在正在以【{job.Name}】的身份执行任务。
目标：{job.Purpose}
约束：{job.Constraints}
工作流：{job.Workflow}";

            return await _aiService.ChatAsync($"{systemPrompt}\n\n任务内容：{prompt}", job.ModelSelectionStrategy);
        }

        public async Task<string> ExecuteJobTaskAsync(string jobId, string prompt, IPluginContext context, string? employeeId = null)
        {
            var job = await _jobService.GetJobAsync(jobId);
            if (job == null)
            {
                return $"❌ 错误：未找到岗位定义 {jobId}";
            }

            // 获取或创建员工实例
            EmployeeInstance? employee = null;
            if (!string.IsNullOrEmpty(employeeId))
            {
                employee = await _employeeService.GetEmployeeAsync(employeeId);
            }
            
            if (employee == null)
            {
                employee = await _employeeService.CreateEmployeeAsync(jobId, employeeId);
            }

            // 更新员工状态为忙碌
            await _employeeService.UpdateEmployeeStateAsync(employee.EmployeeId, "working");

            // 记录任务
            var taskRecord = new TaskRecord
            {
                ExecutionId = Guid.NewGuid(),
                AssigneeId = employee.Id,
                Description = prompt,
                Status = "in_progress",
                StartedAt = DateTime.Now
            };
            var taskId = await _taskRepository.AddAsync(taskRecord);
            taskRecord.Id = taskId;

            try
            {
                // 构建岗位感知的系统提示词
                var jobSystemPrompt = $@"# 岗位职责执行指令
你现在正在以【{job.Name}】的身份执行任务。

## 岗位目标 (Purpose)
{job.Purpose}

## 执行约束 (Constraints)
{job.Constraints}

## 工作流 (Workflow)
请严格遵循以下执行逻辑：
{job.Workflow}

---
当前用户任务请求：
{prompt}";

                // 记录执行步骤
                var step = new TaskStep
                {
                    TaskId = taskId,
                    StepIndex = 0,
                    Name = "MainProcess",
                    Status = "running",
                    InputData = prompt
                };
                var stepId = await _stepRepository.AddAsync(step);
                step.Id = stepId;

                var startTime = DateTime.Now;
                // 使用岗位 Prompt 执行，并尊重模型选择策略
                var result = await _aiService.ChatWithContextAsync(jobSystemPrompt, context, job.ModelSelectionStrategy);
                var duration = (int)(DateTime.Now - startTime).TotalMilliseconds;

                // 记录步骤结束
                step.OutputData = result;
                step.DurationMs = duration;
                
                // 自动化评估
                await _evaluationService.EvaluateStepAsync(step, prompt);
                
                // 更新任务记录
                taskRecord.ResultData = result;
                taskRecord.FinishedAt = DateTime.Now;
                await _evaluationService.EvaluateTaskResultAsync(taskRecord);

                return result;
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "[AgentExecutor] Error executing job task {JobId}", jobId);
                return $"❌ 错误：执行任务时发生异常 - {ex.Message}";
            }
            finally
            {
                // 恢复员工为空闲
                await _employeeService.UpdateEmployeeStateAsync(employee.EmployeeId, "idle");
            }
        }
    }
}


