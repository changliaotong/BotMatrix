namespace BotWorker.Application.Services
{
    public class BrowserMcpHost(IBrowserService browserService) : IMCPHost
    {
        public Task<IEnumerable<MCPTool>> ListToolsAsync(string serverId, CancellationToken ct = default)
        {
            var tools = new List<MCPTool>
            {
                new MCPTool
                {
                    Name = "browser_navigate",
                    Description = "Navigate to a URL and return the text content.",
                    InputSchema = new Dictionary<string, object>
                    {
                        ["type"] = "object",
                        ["properties"] = new Dictionary<string, object>
                        {
                            ["url"] = new { type = "string", description = "The URL to visit" }
                        },
                        ["required"] = new[] { "url" }
                    }
                },
                new MCPTool
                {
                    Name = "browser_screenshot",
                    Description = "Take a full-page screenshot of a URL.",
                    InputSchema = new Dictionary<string, object>
                    {
                        ["type"] = "object",
                        ["properties"] = new Dictionary<string, object>
                        {
                            ["url"] = new { type = "string", description = "The URL to visit" }
                        },
                        ["required"] = new[] { "url" }
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
                var url = arguments["url"].ToString()!;
                switch (toolName)
                {
                    case "browser_navigate":
                        var content = await browserService.NavigateAsync(url);
                        if (content.Length > 10000) content = content.Substring(0, 10000) + "... (truncated)";
                        return Success(content);

                    case "browser_screenshot":
                        var screenshot = await browserService.TakeScreenshotAsync(url);
                        return Success(screenshot);

                    default:
                        return Error($"Unknown tool: {toolName}");
                }
            }
            catch (Exception ex)
            {
                return Error(ex.Message);
            }
        }

        private MCPCallToolResponse Success(string text) => new MCPCallToolResponse { Content = new List<MCPContent> { new MCPContent { Text = text } } };
        private MCPCallToolResponse Error(string text) => new MCPCallToolResponse { IsError = true, Content = new List<MCPContent> { new MCPContent { Text = text } } };
    }
}



