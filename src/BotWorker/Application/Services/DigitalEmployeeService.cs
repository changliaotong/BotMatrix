using System.Threading.Tasks;

namespace BotWorker.Application.Services
{
    public interface IDigitalEmployeeService
    {
        Task<string> ExecuteTaskAsync(string taskDescription);
    }

    public class DigitalEmployeeService(IAIService aiService) : IDigitalEmployeeService
    {
        public async Task<string> ExecuteTaskAsync(string taskDescription)
        {
            return await aiService.ChatAsync($"执行任务: {taskDescription}");
        }
    }
}


