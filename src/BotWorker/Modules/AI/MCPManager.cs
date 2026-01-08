using System;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.Linq;
using System.Threading;
using System.Threading.Tasks;

namespace BotWorker.Services
{
    public class RegisteredServer
    {
        public MCPServerInfo Info { get; set; } = new();
        public IMCPHost Host { get; set; } = null!;
    }

    public class MCPManager : IMcpService
    {
        private readonly ConcurrentDictionary<string, RegisteredServer> _servers = new();

        public void RegisterServer(MCPServerInfo info, IMCPHost host)
        {
            _servers[info.Id] = new RegisteredServer
            {
                Info = info,
                Host = host
            };
        }

        public void UnregisterServer(string serverId)
        {
            _servers.TryRemove(serverId, out _);
        }

        public async Task<IEnumerable<MCPTool>> GetToolsForContextAsync(long userId, long orgId, CancellationToken ct = default)
        {
            var allTools = new List<MCPTool>();
            foreach (var server in _servers.Values)
            {
                bool allowed = server.Info.Scope switch
                {
                    MCPScope.Global => true,
                    MCPScope.Org => server.Info.OwnerId == orgId,
                    MCPScope.User => server.Info.OwnerId == userId,
                    _ => false
                };

                if (allowed)
                {
                    var tools = await server.Host.ListToolsAsync(server.Info.Id, ct);
                    allTools.AddRange(tools);
                }
            }
            return allTools;
        }

        public async Task<MCPCallToolResponse> CallToolAsync(string serverId, string toolName, Dictionary<string, object> arguments, CancellationToken ct = default)
        {
            if (_servers.TryGetValue(serverId, out var server))
            {
                return await server.Host.CallToolAsync(serverId, toolName, arguments, ct);
            }
            return new MCPCallToolResponse { IsError = true, Content = new List<MCPContent> { new MCPContent { Text = $"Server {serverId} not found" } } };
        }
    }
}



