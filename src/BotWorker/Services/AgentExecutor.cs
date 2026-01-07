using System;
using System.Threading.Tasks;
using System.Collections.Generic;

namespace BotWorker.Services
{
    public interface IAgentExecutor
    {
        Task<string> ExecuteAsync(string prompt, Dictionary<string, object>? context = null);
    }

    public class AgentExecutor : IAgentExecutor
    {
        public async Task<string> ExecuteAsync(string prompt, Dictionary<string, object>? context = null)
        {
            // Agent 执行逻辑占位
            return await Task.FromResult($"Executed: {prompt}");
        }
    }
}
