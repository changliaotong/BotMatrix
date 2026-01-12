namespace BotWorker.Application.Services
{
    public class McpInitializationService : IHostedService
    {
        private readonly IMcpService _mcpService;
        private readonly SandboxService _sandboxService;

        public McpInitializationService(IMcpService mcpService, SandboxService sandboxService)
        {
            _mcpService = mcpService;
            _sandboxService = sandboxService;
        }

        public Task StartAsync(CancellationToken cancellationToken)
        {
            if (_mcpService is MCPManager manager)
            {
                // 1. 注册沙盒 MCP 服务
                manager.RegisterServer(new MCPServerInfo
                {
                    Id = "sandbox",
                    Name = "Docker Sandbox",
                    Description = "隔离的 Docker 容器环境，用于执行不安全代码或测试。",
                    Scope = MCPScope.Global
                }, new SandboxMcpHost(_sandboxService));

                // 2. 注册本地开发 MCP 服务
                manager.RegisterServer(new MCPServerInfo
                {
                    Id = "local_dev",
                    Name = "Local Development",
                    Description = "在宿主机上进行受控的开发操作（Git, dotnet 等）。",
                    Scope = MCPScope.Global
                }, new LocalDevMcpHost(AppDomain.CurrentDomain.BaseDirectory));
                
                Console.WriteLine("MCP Servers initialized successfully: sandbox, local_dev");
            }
            return Task.CompletedTask;
        }

        public Task StopAsync(CancellationToken cancellationToken) => Task.CompletedTask;
    }
}
