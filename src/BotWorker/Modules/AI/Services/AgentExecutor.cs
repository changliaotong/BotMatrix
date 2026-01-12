using BotWorker.Modules.AI.Models.Evolution;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Tools;


namespace BotWorker.Modules.AI.Services
{
    public interface IAgentExecutor
    {
        Task<string> ExecuteAsync(string prompt, Dictionary<string, object>? context = null);
        Task<string> ExecuteAgentTaskAsync(string taskId, string staffId, string prompt, IPluginContext context);
        Task<string> ExecuteJobTaskAsync(string jobId, string prompt, IPluginContext context, string? employeeId = null);
        Task<string> ExecuteByJobAsync(string jobId, string prompt); // 增加一个更简单的内部调用接口
    }

    public class AgentExecutor : IAgentExecutor
    {
        private readonly IAIService _aiService;
        private readonly IToolAuditService _auditService;
        private readonly IServiceProvider _serviceProvider;
        private readonly IJobService _jobService;
        private readonly IEmployeeService _employeeService;
        private readonly IEvaluationService _evaluationService;

        public AgentExecutor(
            IAIService aiService, 
            IToolAuditService auditService, 
            IServiceProvider serviceProvider,
            IJobService jobService,
            IEmployeeService employeeService,
            IEvaluationService evaluationService)
        {
            _aiService = aiService;
            _auditService = auditService;
            _serviceProvider = serviceProvider;
            _jobService = jobService;
            _employeeService = employeeService;
            _evaluationService = evaluationService;
        }

        public async Task<string> ExecuteAsync(string prompt, Dictionary<string, object>? context = null)
        {
            // 如果 context 中包含 IPluginContext，则尝试使用更严格的 Agentic 流程
            if (context != null && context.TryGetValue("PluginContext", out var ctx) && ctx is IPluginContext pluginContext)
            {
                var taskId = context.TryGetValue("TaskId", out var tid) ? tid.ToString() : Guid.NewGuid().ToString();
                var staffId = context.TryGetValue("StaffId", out var sid) ? sid.ToString() : "system";
                
                return await ExecuteAgentTaskAsync(taskId!, staffId!, prompt, pluginContext);
            }

            return await _aiService.ChatAsync(prompt);
        }

        /// <summary>
        /// 遵循“数字员工工具接口规范 v1”的执行流程
        /// </summary>
        public async Task<string> ExecuteAgentTaskAsync(string taskId, string staffId, string prompt, IPluginContext context)
        {
            // 1. Planner: 任务分析与规划
            // 在复杂任务中，Planner 负责将任务拆解。目前我们通过系统提示词引导 AI 进入规划状态
            var planningPrompt = $@"
[任务 ID: {taskId}]
[员工 ID: {staffId}]
你现在作为 'Planner' 角色。请分析以下用户需求，并规划执行步骤。
如果需求简单，直接开始执行。
如果需求复杂，请先在心中拆解步骤。

用户需求：{prompt}";

            // 2. Executor: 执行任务 (带审计拦截器)
            // AIService 内部已经通过 DigitalEmployeeToolFilter 实现了 Executor 职责和风险控制
            var result = await _aiService.ChatWithContextAsync(planningPrompt, context);

            // 3. Reviewer: 结果校验与总结
            // 规范要求 Reviewer 判定结果是否合规、是否完成了用户目标
            var reviewPrompt = $@"
你现在作为 'Reviewer' 角色。
请评估以下任务执行结果是否完整且符合预期。
如果结果中包含 'ERROR: 该操作属于高风险行为'，请向用户解释原因并引导其联系管理员审批。

任务需求：{prompt}
执行结果：{result}

请给出最终的用户答复：";

            return await _aiService.ChatAsync(reviewPrompt);
        }

        public async Task<string> ExecuteByJobAsync(string jobId, string prompt)
        {
            var job = await _jobService.GetJobAsync(jobId);
            if (job == null) return $"Error: Job {jobId} not found.";

            var systemPrompt = $@"你现在正在以【{job.Name}】的身份执行任务。
目标：{job.Purpose}
约束：{job.Constraints}
工作流：{job.Workflow}";

            return await _aiService.ChatAsync($"{systemPrompt}\n\n任务内容：{prompt}");
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
            await _employeeService.UpdateEmployeeStateAsync(employee.EmployeeId, "Working");

            // 记录任务
            var taskId = Guid.NewGuid().ToString();
            var taskRecord = new TaskRecord
            {
                TaskId = taskId,
                EmployeeId = employee.EmployeeId,
                InputPayload = prompt,
                Status = "InProgress"
            };
            await taskRecord.SaveAsync();

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

                // 记录执行开始
                var execution = new TaskExecution
                {
                    ExecutionId = Guid.NewGuid().ToString(),
                    TaskId = taskId,
                    StepName = "MainProcess",
                    InputData = prompt,
                    RawPrompt = jobSystemPrompt,
                    StartedAt = DateTime.Now
                };

                // 使用岗位 Prompt 执行
                var result = await _aiService.ChatWithContextAsync(jobSystemPrompt, context);

                // 记录执行结束
                execution.FinishedAt = DateTime.Now;
                execution.RawResponse = result;
                execution.OutputData = result;
                
                // 自动化评估
                await _evaluationService.EvaluateExecutionAsync(execution, prompt);
                
                // 更新任务记录
                taskRecord.ResultOutput = result;
                taskRecord.FinalScore = execution.EvaluationScore;
                await _evaluationService.EvaluateTaskResultAsync(taskRecord);

                return result;
            }
            finally
            {
                // 恢复员工为空闲
                await _employeeService.UpdateEmployeeStateAsync(employee.EmployeeId, "Idle");
            }
        }
    }
}


