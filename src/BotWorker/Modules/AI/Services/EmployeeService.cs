using BotWorker.Modules.AI.Models.Evolution;
using BotWorker.Modules.AI.Interfaces;
using Microsoft.Extensions.Logging;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;

namespace BotWorker.Modules.AI.Services
{
    public class EmployeeService : IEmployeeService
    {
        private readonly ILogger<EmployeeService> _logger;
        private readonly IJobService _jobService;
        private readonly IEmployeeInstanceRepository _employeeRepository;

        public EmployeeService(ILogger<EmployeeService> logger, IJobService jobService, IEmployeeInstanceRepository employeeRepository)
        {
            _logger = logger;
            _jobService = jobService;
            _employeeRepository = employeeRepository;
        }

        public async Task<EmployeeInstance?> GetEmployeeAsync(string employeeId)
        {
            try
            {
                return await _employeeRepository.GetByEmployeeIdAsync(employeeId);
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
                BotId = "system", // 默认系统机器人
                JobId = job.Id,
                Name = job.Name,
                Title = job.Name,
                OnlineStatus = "online",
                State = "idle",
                SalaryTokenLimit = 1000000,
                ExperienceData = "{}"
            };

            var id = await _employeeRepository.AddAsync(instance);
            instance.Id = id;
            return instance;
        }

        public async Task<bool> UpdateEmployeeStateAsync(string employeeId, string state)
        {
            try
            {
                var employee = await _employeeRepository.GetByEmployeeIdAsync(employeeId);
                if (employee == null) return false;
                
                employee.State = state;
                return await _employeeRepository.UpdateAsync(employee);
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "[EmployeeService] Error updating employee state {EmployeeId}", employeeId);
                return false;
            }
        }
    }
}
