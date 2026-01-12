using BotWorker.Modules.AI.Models.Evolution;
using BotWorker.Infrastructure.Persistence.ORM;
using Microsoft.Extensions.Logging;

namespace BotWorker.Modules.AI.Services
{
    public class JobService : IJobService
    {
        private readonly ILogger<JobService> _logger;

        public JobService(ILogger<JobService> logger)
        {
            _logger = logger;
        }

        public async Task<JobDefinition?> GetJobAsync(string jobId, int? version = null)
        {
            try
            {
                var sql = $"SELECT * FROM {JobDefinition.FullName} WHERE JobId = {jobId.Quotes()}";
                if (version.HasValue)
                {
                    sql += $" AND Version = {version.Value}";
                }
                else
                {
                    sql += " AND IsActive = 1 ORDER BY Version DESC";
                }
                
                var jobs = await JobDefinition.QueryListAsync<JobDefinition>(sql);
                return jobs.FirstOrDefault();
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
                var sql = $"SELECT * FROM {JobDefinition.FullName} WHERE IsActive = 1 ORDER BY CreatedAt DESC";
                return await JobDefinition.QueryListAsync<JobDefinition>(sql);
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
                var res = await job.SaveAsync();
                return res > 0;
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "[JobService] Error saving job {JobId}", job.JobId);
                return false;
            }
        }

        public async Task<bool> DeactivateJobAsync(string jobId)
        {
            try
            {
                var sql = $"UPDATE {JobDefinition.FullName} SET IsActive = 0 WHERE JobId = {jobId.Quotes()}";
                var res = await JobDefinition.ExecAsync(sql);
                return res > 0;
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "[JobService] Error deactivating job {JobId}", jobId);
                return false;
            }
        }

        public async Task SeedJobsAsync()
        {
            var defaultJobs = new List<JobDefinition>
            {
                new JobDefinition
                {
                    JobId = "image_refiner",
                    Name = "图像提示词优化专家",
                    Purpose = "将用户简单的描述转化为详细、专业、具有电影感和艺术感的 AI 绘画提示词。",
                    Constraints = "1. 必须包含风格、光影、细节描述。2. 直接输出提示词，不含解释。3. 优先使用中文。",
                    Workflow = "[{\"step\": 1, \"action\": \"分析需求\"}, {\"step\": 2, \"action\": \"风格增强\"}, {\"step\": 3, \"action\": \"输出提示词\"}]"
                },
                new JobDefinition
                {
                    JobId = "code_reviewer",
                    Name = "代码审查工程师",
                    Purpose = "对提交的代码进行安全、性能和规范性审查。",
                    Constraints = "1. 必须指出潜在的内存泄漏风险。2. 检查命名规范。3. 给出优化后的代码示例。",
                    Workflow = "[{\"step\": 1, \"action\": \"静态分析\"}, {\"step\": 2, \"action\": \"性能评估\"}, {\"step\": 3, \"action\": \"生成报告\"}]"
                },
                new JobDefinition
                {
                    JobId = "dev_ra",
                    Name = "需求分析师 (RA)",
                    Purpose = "解析原始需求文档并生成技术规格说明书。",
                    Constraints = "1. 输出必须是严格的 JSON 格式。2. 必须包含功能矩阵和 API 列表。3. 识别出潜在的风险点。",
                    Workflow = "1. 原始需求解析 -> 2. 功能点提取 -> 3. 生成规格说明书"
                },
                new JobDefinition
                {
                    JobId = "dev_architect",
                    Name = "软件架构师 (SA)",
                    Purpose = "根据规格说明书设计项目结构和开发任务。",
                    Constraints = "1. 必须遵循 SOLID 原则。2. 输出项目目录树。3. 拆解具体的待办任务列表。",
                    Workflow = "1. 技术栈选型 -> 2. 目录结构设计 -> 3. 任务拆解"
                },
                new JobDefinition
                {
                    JobId = "dev_coder",
                    Name = "软件开发员 (SD)",
                    Purpose = "根据架构设计编写高质量的源代码。",
                    Constraints = "1. 代码必须包含必要的注释。2. 遵循 Clean Code 规范。3. 确保能够编译通过。",
                    Workflow = "1. 环境初始化 -> 2. 业务逻辑实现 -> 3. 单元测试编写"
                },
                new JobDefinition
                {
                    JobId = "system_upgrader",
                    Name = "系统自我改造专家",
                    Purpose = "分析 BotMatrix 源代码，发现坏味道并进行自动化重构或功能增强。",
                    Constraints = "1. 修改代码前必须先读取相关文件。2. 严禁修改数据库连接字符串等核心密钥。3. 每次重构必须说明理由。",
                    Workflow = "1. 代码库感知 (read_code) -> 2. 架构方案反思 -> 3. 自动化修改 (write_code)"
                }
            };

            foreach (var job in defaultJobs)
            {
                if (await GetJobAsync(job.JobId) == null)
                {
                    await job.SaveAsync();
                    _logger.LogInformation("[JobService] Seeded default job: {JobId}", job.JobId);
                }
            }
        }
    }
}
