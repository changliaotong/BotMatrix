using BotWorker.Modules.AI.Models.Evolution;

namespace BotWorker.Modules.AI.Services
{
    public interface IJobService
    {
        Task<JobDefinition?> GetJobAsync(string jobId, int? version = null);
        Task<List<JobDefinition>> ListJobsAsync();
        Task<bool> SaveJobAsync(JobDefinition job);
        Task<bool> DeactivateJobAsync(string jobId);
        Task SeedJobsAsync();
    }

    public interface IEmployeeService
    {
        Task<EmployeeInstance?> GetEmployeeAsync(string employeeId);
        Task<EmployeeInstance> CreateEmployeeAsync(string jobId, string? employeeId = null);
        Task<bool> UpdateEmployeeStateAsync(string employeeId, string state);
    }
}
