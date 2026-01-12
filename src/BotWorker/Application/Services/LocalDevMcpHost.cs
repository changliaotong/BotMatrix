namespace BotWorker.Application.Services
{
    public class LocalDevMcpHost(string baseDir) : IMCPHost
    {
        private readonly string[] _protectedPaths = { ".git", "config.json", ".env", "id_rsa", "id_rsa.pub" };
        private readonly string[] _allowedCmds = { "dotnet", "git", "ls", "dir", "grep", "cat", "find", "mkdir" };

        public Task<IEnumerable<MCPTool>> ListToolsAsync(string serverId, CancellationToken ct = default)
        {
            var tools = new List<MCPTool>
            {
                new MCPTool
                {
                    Name = "dev_read_file",
                    Description = "Read a file from the project workspace.",
                    InputSchema = new Dictionary<string, object>
                    {
                        ["type"] = "object",
                        ["properties"] = new Dictionary<string, object>
                        {
                            ["path"] = new { type = "string", description = "Relative path to the file" }
                        },
                        ["required"] = new[] { "path" }
                    }
                },
                new MCPTool
                {
                    Name = "dev_write_file",
                    Description = "Write content to a file safely.",
                    InputSchema = new Dictionary<string, object>
                    {
                        ["type"] = "object",
                        ["properties"] = new Dictionary<string, object>
                        {
                            ["path"] = new { type = "string", description = "Relative path to the file" },
                            ["content"] = new { type = "string", description = "Content to write" }
                        },
                        ["required"] = new[] { "path", "content" }
                    }
                }
            };
            return Task.FromResult<IEnumerable<MCPTool>>(tools);
        }

        public Task<IEnumerable<MCPResource>> ListResourcesAsync(string serverId, CancellationToken ct = default) => Task.FromResult(Enumerable.Empty<MCPResource>());
        public Task<IEnumerable<MCPPrompt>> ListPromptsAsync(string serverId, CancellationToken ct = default) => Task.FromResult(Enumerable.Empty<MCPPrompt>());

        public async Task<MCPCallToolResponse> CallToolAsync(string serverId, string toolName, Dictionary<string, object> arguments, CancellationToken ct = default)
        {
            try
            {
                switch (toolName)
                {
                    case "dev_read_file":
                        var readPath = Path.Combine(baseDir, arguments["path"].ToString()!);
                        if (IsProtected(readPath)) return Error("Path is protected");
                        var content = await File.ReadAllTextAsync(readPath, ct);
                        return Success(content);

                    case "dev_write_file":
                        var writePath = Path.Combine(baseDir, arguments["path"].ToString()!);
                        if (IsProtected(writePath)) return Error("Path is protected");
                        var writeContent = arguments["content"].ToString()!;
                        await File.WriteAllTextAsync(writePath, writeContent, ct);
                        return Success("File written successfully");

                    default:
                        return Error($"Unknown tool: {toolName}");
                }
            }
            catch (Exception ex)
            {
                return Error(ex.Message);
            }
        }

        private bool IsProtected(string path)
        {
            return _protectedPaths.Any(p => path.Contains(p));
        }

        private MCPCallToolResponse Success(string text) => new MCPCallToolResponse { Content = new List<MCPContent> { new MCPContent { Text = text } } };
        private MCPCallToolResponse Error(string text) => new MCPCallToolResponse { IsError = true, Content = new List<MCPContent> { new MCPContent { Text = text } } };
    }
}



