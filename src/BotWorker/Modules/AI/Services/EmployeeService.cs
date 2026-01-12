using BotWorker.Modules.AI.Models.Evolution;
using BotWorker.Infrastructure.Persistence.ORM;
using Microsoft.Extensions.Logging;

namespace BotWorker.Modules.AI.Services
{
    public class EmployeeService : IEmployeeService
    {
        private readonly ILogger<EmployeeService> _logger;
        private readonly IJobService _jobService;

        public EmployeeService(ILogger<EmployeeService> logger, IJobService jobService)
        {
            _logger = logger;
            _jobService = jobService;
        }

        public async Task<EmployeeInstance?> GetEmployeeAsync(string employeeId)
        {
            try
            {
                var sql = $"SELECT * FROM {EmployeeInstance.FullName} WHERE EmployeeId = {employeeId.Quotes()}";
                var list = await EmployeeInstance.QueryListAsync<EmployeeInstance>(sql);
                return list.FirstOrDefault();
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "[EmployeeService] Error getting employee {EmployeeId}", employeeId);
                return null;
            }
        }

        public async Task<EmployeeInstance> CreateEmployeeAsync(string jobId, string? employeeId = null)
        {
            var job = await _jobService.GetJobAsync(jobId);
            if (job == null)
            {
                throw new Exception($"Job {jobId} not found");
            }

            var instance = new EmployeeInstance
            {
                EmployeeId = employeeId ?? $"de_{Guid.NewGuid():N}",
                JobId = job.JobId,
                SkillSet = job.Workflow, // 默认技能集可以从 Workflow 或其它字段推导
                PermissionSet = job.Constraints,
                State = "Idle",
                Version = job.Version
            };

            await instance.SaveAsync();
            return instance;
        }

        public async Task<bool> UpdateEmployeeStateAsync(string employeeId, string state)
        {
            try
            {
                var sql = $"UPDATE {EmployeeInstance.FullName} SET State = {state.Quotes()} WHERE EmployeeId = {employeeId.Quotes()}";
                var res = await EmployeeInstance.ExecAsync(sql);
                return res > 0;
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "[EmployeeService] Error updating employee state {EmployeeId}", employeeId);
                return false;
            }
        }
    }
}
