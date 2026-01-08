using System;
using System.Collections.Generic;
using System.Threading.Tasks;

namespace BotWorker.Services
{
    public interface IAIService
    {
        Task<string> ChatAsync(string prompt, string? model = null);
    }

    public class AIService : IAIService
    {
        private readonly IMcpService _mcpService;

        public AIService(IMcpService mcpService)
        {
            _mcpService = mcpService;
        }

        public async Task<string> ChatAsync(string prompt, string? model = null)
        {
            // 基础实现，后续根据需求扩�?
            return $"AI Response to: {prompt} (model: {model ?? "default"})";
        }
    }
}



