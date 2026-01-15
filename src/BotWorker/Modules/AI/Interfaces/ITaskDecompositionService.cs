using System.Collections.Generic;
using System.Threading.Tasks;

namespace BotWorker.Modules.AI.Interfaces
{
    public class SubTaskInfo
    {
        public string Title { get; set; } = string.Empty;
        public string Description { get; set; } = string.Empty;
        public string JobKey { get; set; } = "dev_orchestrator"; // 默认使用开发编排员
        public List<string> Dependencies { get; set; } = new();
    }

    public interface ITaskDecompositionService
    {
        /// <summary>
        /// 将复杂任务拆分为子任务列表
        /// </summary>
        Task<List<SubTaskInfo>> DecomposeAsync(string complexTask);
    }
}
