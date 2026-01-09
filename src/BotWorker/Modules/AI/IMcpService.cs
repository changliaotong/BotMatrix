using System;
using System.Collections.Generic;
using System.Threading;
using System.Threading.Tasks;

namespace BotWorker.Services
{
    public enum MCPScope
    {
        Global,
        Org,
        User
    }

    public class MCPServerInfo
    {
        public string Id { get; set; } = string.Empty;
        public string Name { get; set; } = string.Empty;
        public MCPScope Scope { get; set; } = MCPScope.Global;
        public long OwnerId { get; set; }
    }

    public interface IMcpService
    {
        void RegisterServer(MCPServerInfo info, IMCPHost host);
        void UnregisterServer(string serverId);
        Task<IEnumerable<MCPTool>> GetToolsForContextAsync(long userId, long orgId, CancellationToken ct = default);
        Task<MCPCallToolResponse> CallToolAsync(string serverId, string toolName, Dictionary<string, object> arguments, CancellationToken ct = default);
    }

    public interface IMCPHost
    {
        Task<IEnumerable<MCPTool>> ListToolsAsync(string serverId, CancellationToken ct = default);
        Task<IEnumerable<MCPResource>> ListResourcesAsync(string serverId, CancellationToken ct = default);
        Task<IEnumerable<MCPPrompt>> ListPromptsAsync(string serverId, CancellationToken ct = default);
        Task<MCPCallToolResponse> CallToolAsync(string serverId, string toolName, Dictionary<string, object> arguments, CancellationToken ct = default);
    }

    public class MCPTool
    {
        public string Name { get; set; } = string.Empty;
        public string Description { get; set; } = string.Empty;
        public Dictionary<string, object> InputSchema { get; set; } = new();
        public string ServerId { get; set; } = string.Empty;
    }

    public class MCPResource
    {
        public string URI { get; set; } = string.Empty;
        public string Name { get; set; } = string.Empty;
        public string Description { get; set; } = string.Empty;
        public string MimeType { get; set; } = string.Empty;
    }

    public class MCPPrompt
    {
        public string Name { get; set; } = string.Empty;
        public string Description { get; set; } = string.Empty;
        public List<MCPPromptArgument> Arguments { get; set; } = new();
    }

    public class MCPPromptArgument
    {
        public string Name { get; set; } = string.Empty;
        public string Description { get; set; } = string.Empty;
        public bool Required { get; set; }
    }

    public class MCPCallToolResponse
    {
        public List<MCPContent> Content { get; set; } = new();
        public bool IsError { get; set; }
    }

    public class MCPContent
    {
        public string Type { get; set; } = "text";
        public string Text { get; set; } = string.Empty;
    }
}


