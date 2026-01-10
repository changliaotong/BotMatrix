using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading;
using System.Threading.Tasks;

namespace BotWorker.Application.Services
{
    public class SandboxMcpHost : IMCPHost
    {
        private readonly SandboxService _sandboxService;

        public SandboxMcpHost(SandboxService sandboxService)
        {
            _sandboxService = sandboxService;
        }

        public Task<IEnumerable<MCPTool>> ListToolsAsync(string serverId, CancellationToken ct = default)
        {
            var tools = new List<MCPTool>
            {
                new MCPTool
                {
                    Name = "sandbox_create",
                    Description = "Create a new isolated sandbox environment (Docker container). Returns the sandbox_id.",
                    InputSchema = new Dictionary<string, object>
                    {
                        ["type"] = "object",
                        ["properties"] = new Dictionary<string, object>
                        {
                            ["image"] = new Dictionary<string, object>
                            {
                                ["type"] = "string",
                                ["description"] = "Docker image to use (default: python:3.10-slim)"
                            }
                        }
                    }
                },
                new MCPTool
                {
                    Name = "sandbox_exec",
                    Description = "Execute a shell command inside the sandbox.",
                    InputSchema = new Dictionary<string, object>
                    {
                        ["type"] = "object",
                        ["properties"] = new Dictionary<string, object>
                        {
                            ["sandbox_id"] = new Dictionary<string, object>
                            {
                                ["type"] = "string",
                                ["description"] = "The ID of the sandbox returned by sandbox_create"
                            },
                            ["command"] = new Dictionary<string, object>
                            {
                                ["type"] = "string",
                                ["description"] = "Shell command to execute (e.g., 'ls -la', 'python script.py')"
                            }
                        },
                        ["required"] = new List<string> { "sandbox_id", "command" }
                    }
                },
                new MCPTool
                {
                    Name = "sandbox_write_file",
                    Description = "Write content to a file inside the sandbox.",
                    InputSchema = new Dictionary<string, object>
                    {
                        ["type"] = "object",
                        ["properties"] = new Dictionary<string, object>
                        {
                            ["sandbox_id"] = new Dictionary<string, object>
                            {
                                ["type"] = "string",
                                ["description"] = "The ID of the sandbox"
                            },
                            ["path"] = new Dictionary<string, object>
                            {
                                ["type"] = "string",
                                ["description"] = "Absolute path to the file (e.g., '/workspace/script.py')"
                            },
                            ["content"] = new Dictionary<string, object>
                            {
                                ["type"] = "string",
                                ["description"] = "Text content to write"
                            }
                        },
                        ["required"] = new List<string> { "sandbox_id", "path", "content" }
                    }
                },
                new MCPTool
                {
                    Name = "sandbox_read_file",
                    Description = "Read content from a file inside the sandbox.",
                    InputSchema = new Dictionary<string, object>
                    {
                        ["type"] = "object",
                        ["properties"] = new Dictionary<string, object>
                        {
                            ["sandbox_id"] = new Dictionary<string, object>
                            {
                                ["type"] = "string",
                                ["description"] = "The ID of the sandbox"
                            },
                            ["path"] = new Dictionary<string, object>
                            {
                                ["type"] = "string",
                                ["description"] = "Absolute path to the file"
                            }
                        },
                        ["required"] = new List<string> { "sandbox_id", "path" }
                    }
                },
                new MCPTool
                {
                    Name = "sandbox_destroy",
                    Description = "Destroy the sandbox and release resources.",
                    InputSchema = new Dictionary<string, object>
                    {
                        ["type"] = "object",
                        ["properties"] = new Dictionary<string, object>
                        {
                            ["sandbox_id"] = new Dictionary<string, object>
                            {
                                ["type"] = "string",
                                ["description"] = "The ID of the sandbox"
                            }
                        },
                        ["required"] = new List<string> { "sandbox_id" }
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
                    case "sandbox_create":
                        string? image = arguments.ContainsKey("image") ? arguments["image"]?.ToString() : null;
                        var sb = await _sandboxService.CreateSandboxAsync(image, ct);
                        return new MCPCallToolResponse
                        {
                            Content = new List<MCPContent> { new MCPContent { Text = $"Sandbox created successfully. ID: {sb.ID}" } }
                        };

                    case "sandbox_exec":
                        string id = arguments["sandbox_id"]?.ToString() ?? throw new Exception("missing sandbox_id");
                        string cmd = arguments["command"]?.ToString() ?? throw new Exception("missing command");
                        var (stdout, stderr) = await _sandboxService.ExecInContainerAsync(id, cmd, ct);
                        var output = $"STDOUT:\n{stdout}\n";
                        if (!string.IsNullOrEmpty(stderr)) output += $"\nSTDERR:\n{stderr}";
                        return new MCPCallToolResponse
                        {
                            Content = new List<MCPContent> { new MCPContent { Text = output } }
                        };

                    case "sandbox_write_file":
                        id = arguments["sandbox_id"]?.ToString() ?? throw new Exception("missing sandbox_id");
                        string path = arguments["path"]?.ToString() ?? throw new Exception("missing path");
                        string content = arguments["content"]?.ToString() ?? throw new Exception("missing content");
                        await _sandboxService.WriteFileToContainerAsync(id, path, Encoding.UTF8.GetBytes(content), ct);
                        return new MCPCallToolResponse
                        {
                            Content = new List<MCPContent> { new MCPContent { Text = $"Successfully wrote to {path}" } }
                        };

                    case "sandbox_read_file":
                        id = arguments["sandbox_id"]?.ToString() ?? throw new Exception("missing sandbox_id");
                        path = arguments["path"]?.ToString() ?? throw new Exception("missing path");
                        var fileContent = await _sandboxService.ReadFileFromContainerAsync(id, path, ct);
                        return new MCPCallToolResponse
                        {
                            Content = new List<MCPContent> { new MCPContent { Text = fileContent } }
                        };

                    case "sandbox_destroy":
                        id = arguments["sandbox_id"]?.ToString() ?? throw new Exception("missing sandbox_id");
                        await _sandboxService.DestroySandboxAsync(id, ct);
                        return new MCPCallToolResponse
                        {
                            Content = new List<MCPContent> { new MCPContent { Text = "Sandbox destroyed successfully" } }
                        };

                    default:
                        throw new Exception($"Unknown tool: {toolName}");
                }
            }
            catch (Exception ex)
            {
                return new MCPCallToolResponse
                {
                    IsError = true,
                    Content = new List<MCPContent> { new MCPContent { Text = ex.Message } }
                };
            }
        }
    }
}



