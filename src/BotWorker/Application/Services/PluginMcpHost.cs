using BotWorker.Infrastructure.Communication.OneBot;
using BotWorker.Infrastructure.Utils;

namespace BotWorker.Application.Services
{
    public class PluginMcpHost : IMCPHost
    {
        private readonly PluginManager _pluginManager;
        private readonly IAIService _aiService;
        private readonly II18nService _i18nService;

        public PluginMcpHost(PluginManager pluginManager, IAIService aiService, II18nService i18nService)
        {
            _pluginManager = pluginManager;
            _aiService = aiService;
            _i18nService = i18nService;
        }

        public Task<IEnumerable<MCPTool>> ListToolsAsync(string serverId, CancellationToken ct = default)
        {
            var tools = _pluginManager.Skills.Select(skill => new MCPTool
            {
                Name = skill.Capability.Name,
                Description = skill.Capability.Description + (string.IsNullOrEmpty(skill.Capability.Usage) ? "" : "\n用法: " + skill.Capability.Usage),
                InputSchema = new Dictionary<string, object>
                {
                    { "type", "object" },
                    { "properties", new Dictionary<string, object>
                        {
                            { "args", new Dictionary<string, object>
                                {
                                    { "type", "string" },
                                    { "description", "插件参数，多个参数用空格分隔" }
                                }
                            }
                        }
                    }
                }
            });

            return Task.FromResult(tools);
        }

        public Task<IEnumerable<MCPResource>> ListResourcesAsync(string serverId, CancellationToken ct = default)
        {
            return Task.FromResult(Enumerable.Empty<MCPResource>());
        }

        public Task<IEnumerable<MCPPrompt>> ListPromptsAsync(string serverId, CancellationToken ct = default)
        {
            return Task.FromResult(Enumerable.Empty<MCPPrompt>());
        }

        public async Task<MCPCallToolResponse> CallToolAsync(string serverId, string toolName, Dictionary<string, object> arguments, CancellationToken ct = default)
        {
            var skill = _pluginManager.Skills.FirstOrDefault(s => s.Capability.Name == toolName);
            if (skill == null)
            {
                return new MCPCallToolResponse
                {
                    IsError = true,
                    Content = new List<MCPContent> { new MCPContent { Text = $"未找到插件工�? {toolName}" } }
                };
            }

            try
            {
                // 准备参数
                string[] args = Array.Empty<string>();
                if (arguments.TryGetValue("args", out var argsObj) && argsObj is string argsStr)
                {
                    args = argsStr.Split(new[] { ' ' }, StringSplitOptions.RemoveEmptyEntries);
                }

                // 创建模拟上下�?
                var ev = new OneBotEvent
                {
                    UserIdLong = arguments.TryGetValue("user_id", out var uid) ? uid.ToString().AsLong() : 0,
                    GroupIdLong = arguments.TryGetValue("group_id", out var gid) ? gid.ToString().AsLong() : 0,
                    RawMessage = "" 
                };

                var ctx = new PluginContext(
                    ev, 
                    "mcp", 
                    "system", 
                    _aiService, 
                    _i18nService);
                
                string result = await skill.Handler(ctx, args);

                return new MCPCallToolResponse
                {
                    Content = new List<MCPContent> { new MCPContent { Text = result } }
                };
            }
            catch (Exception ex)
            {
                return new MCPCallToolResponse
                {
                    IsError = true,
                    Content = new List<MCPContent> { new MCPContent { Text = $"执行插件工具出错: {ex.Message}" } }
                };
            }
        }
    }
}


