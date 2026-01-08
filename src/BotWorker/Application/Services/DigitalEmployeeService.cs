using System.Threading.Tasks;

namespace BotWorker.Application.Services
{
    public interface IDigitalEmployeeService
    {
        Task<string> ExecuteTaskAsync(string taskDescription);
    }

    public class DigitalEmployeeService : IDigitalEmployeeService
    {
        private readonly IAIService _aiService;

        public DigitalEmployeeService(IAIService aiService)
        {
            _aiService = aiService;
        }

        public async Task<string> ExecuteTaskAsync(string taskDescription)
        {
            return await _aiService.ChatAsync($"执行任务: {taskDescription}");
        }
    }
}


