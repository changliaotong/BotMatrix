using System;
using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Modules.AI.Models.Evolution;

namespace BotWorker.Modules.AI.Interfaces
{
    public interface IUniversalAgentManager
    {
        /// <summary>
        /// 执行通用的感知-决策-行动循环
        /// </summary>
        /// <param name="jobKey">岗位标识</param>
        /// <param name="initialTask">初始任务描述</param>
        /// <param name="context">插件上下文</param>
        /// <param name="metadata">可选的元数据（如项目路径等）</param>
        /// <returns>执行结果总结</returns>
        Task<string> RunLoopAsync(string jobKey, string initialTask, IPluginContext context, Dictionary<string, string>? metadata = null);
    }
}
