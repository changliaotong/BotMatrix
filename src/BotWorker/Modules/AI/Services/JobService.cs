using BotWorker.Modules.AI.Models.Evolution;
using BotWorker.Modules.AI.Interfaces;
using BotWorker.Modules.AI.Repositories;
using Microsoft.Extensions.Logging;
using System.Collections.Generic;
using System.Linq;
using System;
using System.Threading.Tasks;

namespace BotWorker.Modules.AI.Services
{
    public class JobService : IJobService
    {
        private readonly ILogger<JobService> _logger;
        private readonly IJobDefinitionRepository _jobRepository;

        public JobService(ILogger<JobService> logger, IJobDefinitionRepository jobRepository)
        {
            _logger = logger;
            _jobRepository = jobRepository;
        }

        public async Task<JobDefinition?> GetJobAsync(string jobId, int? version = null)
        {
            try
            {
                // 注意：由于新架构中 job_key 是字符串，Id 是 long
                // 如果传入的是 key，我们用 GetByKeyAsync
                var job = await _jobRepository.GetByKeyAsync(jobId);
                
                if (job != null && version.HasValue && job.Version != version.Value)
                {
                    // 如果需要特定版本，且当前版本不符，可能需要从历史表查（如果实现了的话）
                    // 目前简化处理，只返回当前版本
                }
                
                return job;
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "[JobService] Error getting job {JobId}", jobId);
                return null;
            }
        }

        public async Task<List<JobDefinition>> ListJobsAsync()
        {
            try
            {
                var jobs = await _jobRepository.GetActiveJobsAsync();
                return jobs.ToList();
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "[JobService] Error listing jobs");
                return new List<JobDefinition>();
            }
        }

        public async Task<bool> SaveJobAsync(JobDefinition job)
        {
            try
            {
                if (job.Id > 0)
                {
                    return await _jobRepository.UpdateAsync(job);
                }
                else
                {
                    var id = await _jobRepository.AddAsync(job);
                    job.Id = id;
                    return id > 0;
                }
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "[JobService] Error saving job {JobId}", job.JobKey);
                return false;
            }
        }

        public async Task<bool> DeactivateJobAsync(string jobId)
        {
            try
            {
                var job = await _jobRepository.GetByKeyAsync(jobId);
                if (job == null) return false;
                
                job.IsActive = false;
                return await _jobRepository.UpdateAsync(job);
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "[JobService] Error deactivating job {JobId}", jobId);
                return false;
            }
        }

        public async Task SeedJobsAsync()
        {
            // 1. 预置通用技能
            var defaultSkills = new List<SkillDefinition>
            {               
                new SkillDefinition { SkillKey = "file_read", Name = "读取文件", Description = "读取指定文件的内容", ActionName = "READ", ParameterSchema = "{\"type\": \"string\", \"description\": \"文件路径\"}", IsBuiltin = true },
                new SkillDefinition { SkillKey = "file_write", Name = "写入文件", Description = "将内容写入指定文件", ActionName = "WRITE", ParameterSchema = "{\"type\": \"string\", \"description\": \"文件内容\"}", IsBuiltin = true },
                new SkillDefinition { SkillKey = "list_dir", Name = "列出目录", Description = "列出指定目录下的文件和文件夹", ActionName = "LIST", ParameterSchema = "{\"type\": \"string\", \"description\": \"目录路径\"}", IsBuiltin = true },
                new SkillDefinition { SkillKey = "shell_exec", Name = "执行命令", Description = "在终端执行 Shell 命令", ActionName = "SHELL", ParameterSchema = "{\"type\": \"string\", \"description\": \"命令内容\"}", IsBuiltin = true },
                new SkillDefinition { SkillKey = "git_op", Name = "Git操作", Description = "执行 Git 相关操作", ActionName = "GIT", ParameterSchema = "{\"type\": \"string\", \"description\": \"git命令\"}", IsBuiltin = true },
                new SkillDefinition { SkillKey = "build_check", Name = "编译检查", Description = "执行项目编译并检查错误", ActionName = "BUILD", ParameterSchema = "{\"type\": \"string\", \"description\": \"编译命令\"}", IsBuiltin = true },
                new SkillDefinition { SkillKey = "manage_plan", Name = "计划管理", Description = "制定或更新开发计划", ActionName = "PLAN", ParameterSchema = "{\"type\": \"string\", \"description\": \"计划内容(Markdown格式)\"}", IsBuiltin = true },
                new SkillDefinition 
                { 
                    SkillKey = "python_test", 
                    Name = "Python测试技能", 
                    Description = "用于测试动态 Python 技能的执行", 
                    ActionName = "PYTEST", 
                    ParameterSchema = "{\"type\": \"string\", \"description\": \"测试内容\"}", 
                    IsBuiltin = false,
                    ScriptContent = "import sys\nimport json\n\ndef main():\n    if len(sys.argv) < 2:\n        print(\"Error: No input file provided\")\n        sys.exit(1)\n        \n    input_file = sys.argv[1]\n    with open(input_file, 'r', encoding='utf-8') as f:\n        data = json.load(f)\n        \n    action = data.get('action')\n    target = data.get('target')\n    reason = data.get('reason')\n    \n    result = f\"Python 动态技能执行成功！\\nAction: {action}\\nTarget: {target}\\nReason: {reason}\"\n    print(result)\n\nif __name__ == \"__main__\":\n    main()"
                }
            };
 
             var skillRepo = _jobRepository is PostgresJobDefinitionRepository ? new PostgresSkillDefinitionRepository() : null;
             if (skillRepo != null)
             {
                 foreach (var skill in defaultSkills)
                 {
                     var existing = await skillRepo.GetByKeyAsync(skill.SkillKey);
                     if (existing == null) await skillRepo.AddAsync(skill);
                 }
             }
 
             // 2. 预置岗位
            var defaultJobs = new List<JobDefinition>
            {               
                new JobDefinition
                {
                    JobKey = "image_refiner",
                    Name = "图像提示词优化专家",
                    Purpose = "将用户简单的描述转化为详细、专业、具有电影感和艺术感的 AI 绘画提示词。",
                    Constraints = "[\"必须包含风格、光影、细节描述\", \"直接输出提示词，不含解释\", \"优先使用中文\"]",
                    SystemPrompt = "你是一位专业的 AI 绘画提示词工程师。你的任务是接收用户的简单想法，并将其转化为高质量、充满细节的 Midjourney 或 Stable Diffusion 提示词。你会考虑构图、光照、材质和艺术风格。",
                    ToolSchema = "[]",
                    Workflow = "[{\"step\": 1, \"action\": \"分析需求\"}, {\"step\": 2, \"action\": \"风格增强\"}, {\"step\": 3, \"action\": \"输出提示词\"}]",
                    ModelSelectionStrategy = "random"
                },
                new JobDefinition
                {
                    JobKey = "code_reviewer",
                    Name = "代码审查工程师",
                    Purpose = "对提交的代码进行安全、性能和规范性审查。",
                    Constraints = "[\"必须指出潜在的内存泄漏风险\", \"检查命名规范\", \"给出优化后的代码示例\"]",
                    SystemPrompt = "你是一位资深的软件架构师和代码审查专家。你需要审查用户提交的代码片段或文件，找出其中的 Bug、性能瓶颈、安全隐患以及不符合最佳实践的地方。请以专业、严谨且建设性的态度提供反馈。",
                    ToolSchema = "[\"file_read\"]",
                    Workflow = "[{\"step\": 1, \"action\": \"安全检查\"}, {\"step\": 2, \"action\": \"性能评估\"}, {\"step\": 3, \"action\": \"输出报告\"}]",
                    ModelSelectionStrategy = "random"
                },
                new JobDefinition
                {
                    JobKey = "dev_orchestrator",
                    Name = "开发调度官",
                    Purpose = "作为开发流程的核心，协调需求、架构、编码和测试，根据当前项目状态决定下一步行动。",
                    Constraints = "[\"必须输出合法的 JSON 格式行动建议\", \"在执行任何修改前必须先执行 PLAN 制定步骤\", \"每次 WRITE 或子岗位调用后必须执行 BUILD 检查编译\"]",
                    SystemPrompt = "你是一个全能的数字员工团队领导者。你负责协调复杂的软件开发任务。你需要通过 LIST 和 READ 了解项目，通过 PLAN 规划步骤，然后调用子岗位(如 dev_coder)或直接使用 WRITE 编写代码。编写完代码后，必须调用 BUILD 检查结果，最后使用 GIT 提交。你的目标是实现稳定、可编译的代码交付。",
                    ToolSchema = "[\"list_dir\", \"file_read\", \"file_write\", \"build_check\", \"git_op\", \"manage_plan\"]",
                    Workflow = "[{\"step\": 1, \"action\": \"分析当前状态并制定PLAN\"}, {\"step\": 2, \"action\": \"调度工具或岗位执行任务\"}]",
                    ModelSelectionStrategy = "specified"
                },
                new JobDefinition
                {
                    JobKey = "dev_git",
                    Name = "版本管理员",
                    Purpose = "管理项目的 Git 仓库，处理提交、分支和冲突。",
                    Constraints = "[\"输出标准的 git 命令\", \"每个 commit 必须有清晰的描述\"]",
                    SystemPrompt = "你是一个精通 Git 工作流的版本管理专家。你负责确保代码仓库的整洁和提交记录的规范。你会根据开发进度执行 commit、branch、merge 等操作，并妥善处理可能出现的冲突。",
                    ToolSchema = "[\"git_op\"]",
                    Workflow = "[{\"step\": 1, \"action\": \"检查状态\"}, {\"step\": 2, \"action\": \"执行提交/分支操作\"}]",
                    ModelSelectionStrategy = "specified"
                },
                new JobDefinition
                {
                    JobKey = "dev_ra",
                    Name = "需求分析专家 (RA)",
                    Purpose = "将模糊的自然语言需求转化为结构化的技术规格说明书。",
                    Constraints = "[\"输出必须是合法的 JSON\", \"必须定义功能点、非功能性要求和边界条件\", \"识别关键技术栈\"]",
                    SystemPrompt = "你是一位擅长沟通与分析的需求分析师。你能精准捕捉用户模糊意图中的核心价值，并将其转化为开发团队可理解、可执行的技术需求规格。你注重细节，善于发现需求中的矛盾点和遗漏点。",
                    ToolSchema = "[\"file_read\", \"file_write\"]",
                    Workflow = "[{\"step\": 1, \"action\": \"需求提取\"}, {\"step\": 2, \"action\": \"边界定义\"}, {\"step\": 3, \"action\": \"JSON 规格化\"}]",
                    ModelSelectionStrategy = "specified"
                },
                new JobDefinition
                {
                    JobKey = "dev_architect",
                    Name = "系统架构师 (SA)",
                    Purpose = "根据技术规格说明书设计系统架构和项目目录结构。",
                    Constraints = "[\"输出包含 tasks 列表的 JSON\", \"每个 task 包含 fileName 和 description\", \"遵循领域驱动设计 (DDD) 或其他标准架构\"]",
                    SystemPrompt = "你是一位经验丰富的系统架构师。你擅长设计可扩展、高可用、低耦合的 system 架构。你会根据需求选择最合适的技术栈和设计模式，并规划出清晰的代码组织结构。",
                    ToolSchema = "[\"file_read\", \"file_write\", \"list_dir\"]",
                    Workflow = "[{\"step\": 1, \"action\": \"模块划分\"}, {\"step\": 2, \"action\": \"接口设计\"}, {\"step\": 3, \"action\": \"生成文件清单\"}]",
                    ModelSelectionStrategy = "specified"
                },
                new JobDefinition
                {
                    JobKey = "dev_coder",
                    Name = "软件开发工程师 (SD)",
                    Purpose = "根据功能描述和技术规格编写高质量的源代码。",
                    Constraints = "[\"代码必须符合 clean code 原则\", \"包含必要的注释\", \"如果是直接编写文件，必须使用 WRITE 行动并在回复中包含 ``` 代码块\"]",
                    SystemPrompt = "你是一位追求卓越的代码极客。你精通多种编程语言和框架。你的任务是实现具体的代码逻辑。你可以直接使用 WRITE 技能来创建或修改文件。请确保你的回复中包含完整的代码内容，并包裹在 Markdown 代码块中。",
                    ToolSchema = "[\"file_read\", \"file_write\"]",
                    Workflow = "[{\"step\": 1, \"action\": \"代码实现\"}, {\"step\": 2, \"action\": \"自测验证\"}]",
                    ModelSelectionStrategy = "random"
                },
                new JobDefinition
                {
                    JobKey = "dev_tester",
                    Name = "测试工程师 (QA)",
                    Purpose = "编写单元测试或对代码进行功能性验证。",
                    Constraints = "[\"识别代码中的逻辑漏洞\", \"提供测试覆盖率建议\", \"如果发现问题，必须以 FAIL: 开头\"]",
                    SystemPrompt = "你是一位目光敏锐的质量保证专家。你的使命是发现代码中任何可能导致崩溃、安全漏洞或逻辑错误的问题。你会编写详尽的测试用例，并在代码交付前进行严格的验证。",
                    ToolSchema = "[\"file_read\", \"shell_exec\"]",
                    Workflow = "[{\"step\": 1, \"action\": \"测试用例设计\"}, {\"step\": 2, \"action\": \"执行验证\"}]",
                    ModelSelectionStrategy = "random"
                }
            };

            foreach (var job in defaultJobs)
            {
                var existing = await _jobRepository.GetByKeyAsync(job.JobKey);
                if (existing == null)
                {
                    await SaveJobAsync(job);
                }
                else
                {
                    // 更新现有记录，确保新字段生效
                    existing.Name = job.Name;
                    existing.Purpose = job.Purpose;
                    existing.Constraints = job.Constraints;
                    existing.SystemPrompt = job.SystemPrompt;
                    existing.ToolSchema = job.ToolSchema;
                    existing.Workflow = job.Workflow;
                    existing.ModelSelectionStrategy = job.ModelSelectionStrategy;
                    await SaveJobAsync(existing);
                }
            }
        }
    }
}
