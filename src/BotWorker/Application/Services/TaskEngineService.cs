namespace BotWorker.Application.Services
{
    public interface ITaskEngineService
    {
        Task<string> RunTaskAsync(string taskId, Dictionary<string, object> parameters);
    }

    public class TaskEngineService : ITaskEngineService
    {
        public async Task<string> RunTaskAsync(string taskId, Dictionary<string, object> parameters)
        {
            return await Task.FromResult($"Task {taskId} completed");
        }
    }
}


