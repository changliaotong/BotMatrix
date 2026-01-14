using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Text.Json;
using System.Threading.Tasks;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Models.Evolution;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.Logging;

namespace BotWorker.Modules.AI.Services
{
    public class UniversalAgentManager : IUniversalAgentManager
    {
        private readonly IAIService _aiService;
        private readonly IJobService _jobService;
        private readonly ISkillService _skillService;
        private readonly ISkillDefinitionRepository _skillRepository;
        private readonly ITaskStepRepository _stepRepository;
        private readonly IConfiguration _configuration;
        private readonly ILogger<UniversalAgentManager> _logger;

        public UniversalAgentManager(
            IAIService aiService, 
            IJobService jobService, 
            ISkillService skillService,
            ISkillDefinitionRepository skillRepository,
            ITaskStepRepository stepRepository,
            IConfiguration configuration,
            ILogger<UniversalAgentManager> logger)
        {
            _aiService = aiService;
            _jobService = jobService;
            _skillService = skillService;
            _skillRepository = skillRepository;
            _stepRepository = stepRepository;
            _configuration = configuration;
            _logger = logger;
        }

        public async Task<string> RunLoopAsync(string jobKey, string initialTask, IPluginContext context, Dictionary<string, string>? metadata = null)
        {
            var job = await _jobService.GetJobAsync(jobKey);
            if (job == null)
            {
                return $"Error: Job {jobKey} not found.";
            }

            // 生成隔离的工作区路径
            var workspaceRoot = _configuration["WorkspaceRoot"] ?? Path.Combine(Directory.GetCurrentDirectory(), "BotWorkspaces");
            var tenantId = context.GroupId ?? "default_tenant";
            var userId = context.UserId ?? "default_user";
            var sessionId = Guid.NewGuid().ToString("N").Substring(0, 8); // 简短会话ID
            
            // 尝试从 metadata 中获取 TaskId，如果没有则使用 sessionId
            var taskId = metadata?.GetValueOrDefault("TaskId") ?? sessionId;
            var projectPath = Path.Combine(workspaceRoot, tenantId, userId, taskId);
            
            bool isNewProject = !Directory.Exists(projectPath);
            if (isNewProject)
            {
                Directory.CreateDirectory(projectPath);
                _logger.LogInformation("[UniversalAgent] Created isolated workspace: {Path}", projectPath);
                
                // 自动初始化 Git
                try {
                    await _skillService.ExecuteSkillAsync("GIT", "git init", "初始化项目仓库", new Dictionary<string, string> { ["ProjectPath"] = projectPath });
                    await _skillService.ExecuteSkillAsync("WRITE", ".gitignore", "初始化 gitignore", new Dictionary<string, string> { 
                        ["ProjectPath"] = projectPath,
                        ["Content"] = "bin/\nobj/\n.vs/\n*.log\n"
                    });
                } catch (Exception ex) {
                    _logger.LogWarning("[UniversalAgent] Git init failed: {Message}", ex.Message);
                }
            }

            var agentMetadata = metadata ?? new Dictionary<string, string>();
            agentMetadata["ProjectPath"] = projectPath;
            agentMetadata["TenantId"] = tenantId;
            agentMetadata["UserId"] = userId;
            agentMetadata["TaskId"] = taskId;

            // 获取岗位关联的技能详情
            var skillKeys = JsonSerializer.Deserialize<List<string>>(job.ToolSchema) ?? new List<string>();
            var skills = await _skillRepository.GetByKeysAsync(skillKeys);

            _logger.LogInformation("[UniversalAgent] Starting loop for job {JobKey} with {SkillCount} tools", jobKey, skills.Count());

            var state = new AgentState
            {
                InitialTask = initialTask,
                Metadata = agentMetadata,
                History = new List<string>(),
                Files = new Dictionary<string, string>()
            };

            int maxSteps = 20; // 增加步数上限
            int currentStep = 0;
            string finalResult = "任务超时或未完成";

            while (currentStep < maxSteps)
            {
                currentStep++;
                _logger.LogInformation("[UniversalAgent] {JobKey} Step {Step}/{Max}", jobKey, currentStep, maxSteps);

                // 强化：如果是第一步，要求 Agent 必须先写规划
                if (currentStep == 1 && !state.Files.ContainsKey("plan.md"))
                {
                    state.InitialTask = "请首先根据目标，在工作区根目录创建一个 plan.md 文件，详细列出执行计划。\n目标内容：" + initialTask;
                }

                var prompt = BuildPrompt(job, skills, state);
                var response = await _aiService.ChatWithContextAsync(prompt, context, job.ModelSelectionStrategy);
                _logger.LogInformation("[UniversalAgent] Raw response: {Response}", response);

                var decision = ParseDecision(response);
                _logger.LogInformation("[UniversalAgent] {JobKey} Decision: {Action} on {Target} ({Reason})", 
                    jobKey, decision.Action, decision.Target, decision.Reason);

                // 如果有 TaskId，记录详细步骤
                if (long.TryParse(taskId, out var tid))
                {
                    var stepName = $"{decision.Action}: {decision.Target}";
                    if (stepName.Length > 100) stepName = stepName.Substring(0, 97) + "...";

                    await _stepRepository.AddAsync(new TaskStep
                    {
                        TaskId = tid,
                        StepIndex = currentStep,
                        Name = stepName,
                        InputData = JsonSerializer.Serialize(new { Reason = decision.Reason, Prompt = prompt }),
                        Status = "executing",
                        CreatedAt = DateTime.Now
                    });
                }

                if (decision.Action == "DONE")
                {
                    finalResult = decision.Reason;
                    break;
                }

                var observation = await ExecuteActionAsync(decision, state, context);
                state.LastObservation = observation;
                state.History.Add($"Step {currentStep}: {decision.Action} {decision.Target} -> {observation}");

                // 自动化：执行成功后自动 Git Commit
                if (!observation.StartsWith("错误") && decision.Action.ToUpper() != "READ" && decision.Action.ToUpper() != "LIST" && decision.Action.ToUpper() != "DONE")
                {
                    try {
                        var commitMsg = $"Step {currentStep}: {decision.Action} {decision.Target}";
                        if (commitMsg.Length > 100) commitMsg = commitMsg.Substring(0, 97) + "...";
                        
                        // 先尝试 add
                        await _skillService.ExecuteSkillAsync("GIT", "git add .", "添加变更", agentMetadata);
                        // 再尝试 commit，如果没有任何变更 commit 会失败，我们忽略它
                        await _skillService.ExecuteSkillAsync("GIT", $"git commit -m \"{commitMsg}\"", "自动提交", agentMetadata);
                    } catch { /* 忽略 Git 错误，例如没有变更可提交 */ }
                }

                // 更新步骤结果
                if (long.TryParse(taskId, out var tid2))
                {
                    var steps = await _stepRepository.GetByTaskIdAsync(tid2);
                    var currentTaskStep = steps.FirstOrDefault(s => s.StepIndex == currentStep);
                    if (currentTaskStep != null)
                    {
                        currentTaskStep.OutputData = JsonSerializer.Serialize(new { result = observation });
                        currentTaskStep.Status = "completed";
                        currentTaskStep.UpdatedAt = DateTime.Now;
                        await _stepRepository.UpdateAsync(currentTaskStep);
                    }
                }
            }

            return finalResult;
        }

        private string BuildPrompt(JobDefinition job, IEnumerable<SkillDefinition> skills, AgentState state)
        {
            var sb = new System.Text.StringBuilder();
            sb.AppendLine($"# {job.Name} 执行指令 (Autonomous Mode)");
            sb.AppendLine();
            sb.AppendLine("## 你的身份");
            sb.AppendLine(job.SystemPrompt);
            sb.AppendLine();
            sb.AppendLine("## 核心目标");
            sb.AppendLine(job.Purpose);
            sb.AppendLine();
            sb.AppendLine("## 自动化行为规范 (Manus Protocol)");
            sb.AppendLine("1. **自主规划**: 必须在工作区维护 plan.md，记录已完成和待办事项。");
            sb.AppendLine("2. **自给自足**: 遇到错误（如缺少依赖、编译失败）时，应尝试通过命令行工具自行解决。");
            sb.AppendLine("3. **增量开发**: 每次修改后应通过 BUILD 或 COMMAND 验证。");
            sb.AppendLine("4. **版本控制**: 你的所有文件变更都会被自动 git commit，请确保代码逻辑清晰。");
            sb.AppendLine();
            sb.AppendLine("## 执行约束");
            sb.AppendLine(job.Constraints);
            sb.AppendLine();
            sb.AppendLine("## 可用工具 (Tools)");
            if (!skills.Any())
            {
                sb.AppendLine("无可用特定工具。使用 COMMAND 执行系统命令。");
            }
            foreach (var skill in skills)
            {
                sb.AppendLine($"- {skill.ActionName}: {skill.Description}");
                sb.AppendLine($"  参数格式: {skill.ParameterSchema}");
            }
            sb.AppendLine();
            sb.AppendLine("## 初始任务");
            sb.AppendLine(state.InitialTask);
            sb.AppendLine();
            sb.AppendLine("## 当前工作区状态");
            sb.AppendLine(state.GetSummary());
            sb.AppendLine();
            sb.AppendLine("## 决策要求");
            sb.AppendLine("请分析当前进度，决定下一步行动。");
            sb.AppendLine("输出 JSON 格式：{\"action\": \"工具名\", \"target\": \"目标/参数\", \"reason\": \"原因\", \"content\": \"(可选) 写入内容\"}");
            sb.AppendLine("完成所有目标后，使用 action: \"DONE\"。");

            return sb.ToString();
        }

        private async Task<string> ExecuteActionAsync(AgentDecision decision, AgentState state, IPluginContext context)
        {
            try
            {
                // 准备技能执行所需的元数据
                var metadata = new Dictionary<string, string>(state.Metadata);
                if (!string.IsNullOrEmpty(decision.Content))
                {
                    metadata["Content"] = decision.Content;
                }

                // 优先检查是否是嵌套的岗位调用
                var nestedJob = await _jobService.GetJobAsync(decision.Action.ToLower());
                if (nestedJob != null && nestedJob.JobKey != "dev_orchestrator")
                {
                    _logger.LogInformation("[UniversalAgent] Nesting call to job {JobKey}", nestedJob.JobKey);
                    
                    var nestedPrompt = $@"目标：{decision.Target}
任务说明：{decision.Reason}
内容参数：{decision.Content ?? "无"}
项目上下文：
{state.GetContextForFile(decision.Target)}";

                    var result = await _aiService.ChatWithContextAsync(BuildJobPrompt(nestedJob, nestedPrompt), context, nestedJob.ModelSelectionStrategy);
                    
                    // 如果嵌套调用返回了内容且行动涉及写入，将其缓存
                    if (decision.Action.ToUpper().Contains("CODER") || decision.Action.ToUpper().Contains("WRITE"))
                    {
                        state.Files[decision.Target] = result;
                        
                        // 实际上，这里我们可以选择自动写入文件，或者由 AI 在下一步显式调用 WRITE
                        // 为了 MVP 尽快见效，我们在这里直接调用一次 WRITE 技能
                        await _skillService.ExecuteSkillAsync("WRITE", decision.Target, decision.Reason, new Dictionary<string, string>(metadata) { ["Content"] = result });
                        return $"已通过 {nestedJob.Name} 完成处理并自动写入：{decision.Target}";
                    }

                    return $"岗位 {nestedJob.Name} 执行结果：\n{result}";
                }

                // 使用通用的技能服务执行行动
                var observation = await _skillService.ExecuteSkillAsync(decision.Action, decision.Target, decision.Reason, metadata);
                
                // 特殊处理：如果读取或写入了文件，更新状态缓存以便后续步骤使用上下文
                var cmd = decision.Action.ToUpper();
                if (cmd == "READ" && !observation.StartsWith("错误"))
                {
                    state.Files[decision.Target] = observation;
                }
                else if (cmd == "WRITE")
                {
                    state.Files[decision.Target] = decision.Content ?? decision.Reason;
                }

                return observation;
            }
            catch (Exception ex)
            {
                return $"执行行动时发生错误：{ex.Message}";
            }
        }

        private string BuildJobPrompt(JobDefinition job, string prompt)
        {
            return $@"你现在正在以【{job.Name}】的身份执行子任务。
职责：{job.Purpose}
约束：{job.Constraints}
系统背景：{job.SystemPrompt}

待处理内容：
{prompt}";
        }

        private AgentDecision ParseDecision(string content)
        {
            try
            {
                // 1. 尝试解析 JSON 部分
                var jsonMatch = System.Text.RegularExpressions.Regex.Match(content, @"\{.*\}", System.Text.RegularExpressions.RegexOptions.Singleline);
                AgentDecision decision;
                
                if (jsonMatch.Success)
                {
                    using var doc = JsonDocument.Parse(jsonMatch.Value);
                    var root = doc.RootElement;
                    
                    decision = new AgentDecision();
                    if (root.TryGetProperty("action", out var actionProp)) decision.Action = actionProp.GetString() ?? "DONE";
                    if (root.TryGetProperty("target", out var targetProp))
                    {
                        decision.Target = targetProp.ValueKind == JsonValueKind.String ? targetProp.GetString() ?? "" : targetProp.GetRawText();
                    }
                    if (root.TryGetProperty("reason", out var reasonProp)) decision.Reason = reasonProp.GetString() ?? "";
                    if (root.TryGetProperty("content", out var contentProp)) decision.Content = contentProp.GetString();
                    
                    // 额外处理：如果 Target 中包含换行符，且 Action 是 WRITE，说明模型把内容错放到了 Target 里
                    if (decision.Action.ToUpper() == "WRITE" && decision.Target.Contains("\n"))
                    {
                        var lines = decision.Target.Split(new[] { '\n', '\r' }, StringSplitOptions.RemoveEmptyEntries);
                        if (lines.Length > 0)
                        {
                            decision.Target = lines[0].Trim();
                            if (string.IsNullOrEmpty(decision.Content))
                            {
                                decision.Content = string.Join("\n", lines.Skip(1)).Trim();
                            }
                        }
                    }
                }
                else
                {
                    decision = new AgentDecision { Action = "UNKNOWN", Reason = "解析 JSON 失败" };
                }

                // 2. 如果是 WRITE 行动，尝试提取内容块 (``` ... ```)
                if (decision.Action.ToUpper() == "WRITE" || decision.Action.ToUpper().Contains("CODER"))
                {
                    // 支持多种代码块格式，包括带语言标识的和不带的
                    var codeMatch = System.Text.RegularExpressions.Regex.Match(content, @"```(?:\w+)?\s*\n?(.*?)```", System.Text.RegularExpressions.RegexOptions.Singleline);
                    if (codeMatch.Success)
                    {
                        decision.Content = codeMatch.Groups[1].Value.Trim();
                    }
                    else if (string.IsNullOrEmpty(decision.Content))
                    {
                        // 兜底方案：如果 JSON 里没有 content，也没有代码块，但有明显的代码特征
                        if (content.Contains("{") && content.Contains("}") && (content.Contains("using ") || content.Contains("import ")))
                        {
                             // 提取 JSON 之后的所有文本作为 content
                             var index = content.LastIndexOf('}');
                             if (index > 0 && index < content.Length - 1)
                             {
                                 decision.Content = content.Substring(index + 1).Trim();
                             }
                        }
                        
                        if (string.IsNullOrEmpty(decision.Content))
                        {
                            decision.Content = decision.Reason;
                        }
                    }
                }

                return decision;
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "[UniversalAgent] ParseDecision Error");
                return new AgentDecision { Action = "UNKNOWN", Reason = "解析决策发生异常：" + ex.Message };
            }
        }

        private class AgentState
        {
            public string InitialTask { get; set; } = string.Empty;
            public Dictionary<string, string> Metadata { get; set; } = new();
            public List<string> History { get; set; } = new();
            public Dictionary<string, string> Files { get; set; } = new();
            public string? LastObservation { get; set; }

            public string GetSummary()
            {
                return JsonSerializer.Serialize(new
                {
                    FileList = Files.Keys.ToList(),
                    HistoryCount = History.Count,
                    RecentHistory = History.TakeLast(3).ToList(),
                    LastObservation = LastObservation?.Length > 500 ? LastObservation.Substring(0, 500) + "..." : LastObservation
                }, new JsonSerializerOptions { WriteIndented = true });
            }

            public string GetContextForFile(string fileName)
            {
                var relevantFiles = Files.Where(f => f.Key != fileName).ToList();
                var context = new System.Text.StringBuilder();
                foreach (var file in relevantFiles)
                {
                    context.AppendLine($"--- File: {file.Key} ---");
                    if (file.Value.Length > 2000)
                    {
                        context.AppendLine(file.Value.Substring(0, 1000));
                        context.AppendLine("... [content truncated] ...");
                        context.AppendLine(file.Value.Substring(file.Value.Length - 1000));
                    }
                    else
                    {
                        context.AppendLine(file.Value);
                    }
                    context.AppendLine();
                }
                return context.ToString();
            }
        }

        private class AgentDecision
        {
            public string Action { get; set; } = "DONE";
            public string Target { get; set; } = "";
            public string Reason { get; set; } = "";
            public string? Content { get; set; } // 用于存储大段代码或文本内容
        }
    }
}
