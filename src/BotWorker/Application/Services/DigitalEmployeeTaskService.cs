using System;
using System.Threading.Tasks;
using System.Collections.Generic;

namespace BotWorker.Services
{
    public interface IDigitalEmployeeTaskService
    {
        Task<string> CreateTaskAsync(string employeeId, string prompt);
        Task<string> GetTaskStatusAsync(string taskId);
    }

    public class DigitalEmployeeTaskService : IDigitalEmployeeTaskService
    {
        public async Task<string> CreateTaskAsync(string employeeId, string prompt)
        {
            var taskId = Guid.NewGuid().ToString();
            // 任务创建逻辑
            return await Task.FromResult(taskId);
        }

        public async Task<string> GetTaskStatusAsync(string taskId)
        {
            return await Task.FromResult("Running");
        }
    }
}


