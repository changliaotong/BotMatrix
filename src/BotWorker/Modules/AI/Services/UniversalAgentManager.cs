using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Text.Json;
using System.Threading.Tasks;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Models.Evolution;
using Microsoft.Extensions.Logging;

namespace BotWorker.Modules.AI.Services
{
    public class UniversalAgentManager : IUniversalAgentManager
    {
        private readonly IAIService _aiService;
        private readonly IJobService _jobService;
        private readonly ISkillService _skillService;
        private readonly ISkillDefinitionRepository _skillRepository;
        private readonly ILogger<UniversalAgentManager> _logger;

        public UniversalAgentManager(
            IAIService aiService, 
            IJobService jobService, 
            ISkillService skillService,
            ISkillDefinitionRepository skillRepository,
            ILogger<UniversalAgentManager> logger)
        {
            _aiService = aiService;
            _jobService = jobService;
            _skillService = skillService;
            _skillRepository = skillRepository;
            _logger = logger;
        }

        public async Task<string> RunLoopAsync(string jobKey, string initialTask, IPluginContext context, Dictionary<string, string>? metadata = null)
        {
            var job = await _jobService.GetJobAsync(jobKey);
            if (job == null)
            {
                return $"Error: Job {jobKey} not found.";
            }

            // 获取岗位关联的技能详情
            var skillKeys = JsonSerializer.Deserialize<List<string>>(job.ToolSchema) ?? new List<string>();
            var skills = await _skillRepository.GetByKeysAsync(skillKeys);

            _logger.LogInformation("[UniversalAgent] Starting loop for job {JobKey} with {SkillCount} tools", jobKey, skills.Count());

            var state = new AgentState
            {
                InitialTask = initialTask,
                Metadata = metadata ?? new Dictionary<string, string>(),
                History = new List<string>(),
                Files = new Dictionary<string, string>()
            };

            int maxSteps = 15;
            int currentStep = 0;
            string finalResult = "任务超时或未完成";

            while (currentStep < maxSteps)
            {
                currentStep++;
                _logger.LogInformation("[UniversalAgent] {JobKey} Step {Step}/{Max}", jobKey, currentStep, maxSteps);

                var prompt = BuildPrompt(job, skills, state);
                var response = await _aiService.ChatWithContextAsync(prompt, context, job.ModelSelectionStrategy);

                var decision = ParseDecision(response);
                _logger.LogInformation("[UniversalAgent] {JobKey} Decision: {Action} on {Target} ({Reason})", 
                    jobKey, decision.Action, decision.Target, decision.Reason);

                if (decision.Action == "DONE")
                {
                    finalResult = decision.Reason;
                    break;
                }

                var observation = await ExecuteActionAsync(decision, state, context);
                state.LastObservation = observation;
                state.History.Add($"Step {currentStep}: {decision.Action} {decision.Target} -> {observation}");
            }

            return finalResult;
        }

        private string BuildPrompt(JobDefinition job, IEnumerable<SkillDefinition> skills, AgentState state)
        {
            var sb = new System.Text.StringBuilder();
            sb.AppendLine($"# {job.Name} 执行指令");
            sb.AppendLine();
            sb.AppendLine("## 你的身份");
            sb.AppendLine(job.SystemPrompt);
            sb.AppendLine();
            sb.AppendLine("## 核心目标");
            sb.AppendLine(job.Purpose);
            sb.AppendLine();
            sb.AppendLine("## 执行约束");
            sb.AppendLine(job.Constraints);
            sb.AppendLine();
            sb.AppendLine("## 可用工具 (Tools)");
            if (!skills.Any())
            {
                sb.AppendLine("无可用特定工具。");
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
            sb.AppendLine("## 当前状态总结");
            sb.AppendLine(state.GetSummary());
            sb.AppendLine();
            sb.AppendLine("## 请决定下一步行动");
            sb.AppendLine("必须输出合法的 JSON 格式：{\"action\": \"工具名\", \"target\": \"目标/参数\", \"reason\": \"原因\"}");
            sb.AppendLine("如果任务已完成，请使用 action: \"DONE\"。");

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
                    decision = JsonSerializer.Deserialize<AgentDecision>(jsonMatch.Value, new JsonSerializerOptions { PropertyNameCaseInsensitive = true }) ?? new AgentDecision();
                }
                else
                {
                    decision = new AgentDecision { Action = "UNKNOWN", Reason = "解析 JSON 失败" };
                }

                // 2. 如果是 WRITE 行动，尝试提取内容块 (``` ... ```)
                if (decision.Action.ToUpper() == "WRITE" || decision.Action.ToUpper().Contains("CODER"))
                {
                    var codeMatch = System.Text.RegularExpressions.Regex.Match(content, @"```(?:\w+)?\n?(.*?)```", System.Text.RegularExpressions.RegexOptions.Singleline);
                    if (codeMatch.Success)
                    {
                        decision.Content = codeMatch.Groups[1].Value.Trim();
                    }
                    else if (string.IsNullOrEmpty(decision.Content))
                    {
                        // 如果没有代码块，但 JSON 里也没有 content，则尝试从 reason 提取（兼容模式）
                        decision.Content = decision.Reason;
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
