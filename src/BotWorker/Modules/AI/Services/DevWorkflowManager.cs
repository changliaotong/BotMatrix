using BotWorker.Modules.AI.Interfaces;
using Microsoft.Extensions.Logging;
using System.Text.Json;

namespace BotWorker.Modules.AI.Services
{
    public class DevWorkflowManager : IDevWorkflowManager
    {
        private readonly IUniversalAgentManager _agentManager;
        private readonly ILogger<DevWorkflowManager> _logger;

        public DevWorkflowManager(IUniversalAgentManager agentManager, ILogger<DevWorkflowManager> logger)
        {
            _agentManager = agentManager;
            _logger = logger;
        }

        public async Task<bool> StartDevProjectAsync(string requirementDoc, string projectPath)
        {
            _logger.LogInformation("[DevWorkflow] Delegating to UniversalAgentManager for project in {Path}", projectPath);

            var metadata = new Dictionary<string, string>
            {
                { "ProjectPath", projectPath }
            };

            // 使用通用的感知-决策-行动循环
            // 这里我们假定有一个通用的 IPluginContext，或者在某些场景下可以为 null
            // 由于 StartDevProjectAsync 目前没有 context 参数，我们需要考虑如何获取或模拟一个
            // 暂时传入 null，或者修改接口。但为了保持兼容性，我们先看看 UniversalAgentManager 是否处理 null。
            
            var result = await _agentManager.RunLoopAsync("dev_orchestrator", requirementDoc, null!, metadata);
            
            _logger.LogInformation("[DevWorkflow] Result: {Result}", result);
            return true;
        }


    }
}
