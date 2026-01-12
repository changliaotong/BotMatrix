namespace BotWorker.Modules.AI.Interfaces
{
    public interface IEvolutionService
    {
        /// <summary>
        /// 针对特定岗位进行进化分析，并决定是否更新岗位定义
        /// </summary>
        Task<bool> EvolveJobAsync(string jobId);
        
        /// <summary>
        /// 执行所有待进化的岗位
        /// </summary>
        Task EvolveAllJobsAsync();
    }
}
