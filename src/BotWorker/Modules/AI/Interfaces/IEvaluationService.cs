using BotWorker.Modules.AI.Models.Evolution;

namespace BotWorker.Modules.AI.Interfaces
{
    public interface IEvaluationService
    {
        /// <summary>
        /// 对单次执行步骤进行评估
        /// </summary>
        Task<bool> EvaluateStepAsync(TaskStep step, string taskPrompt);
        
        /// <summary>
        /// 对整个任务结果进行最终评估
        /// </summary>
        Task<bool> EvaluateTaskResultAsync(TaskRecord task);
    }
}
