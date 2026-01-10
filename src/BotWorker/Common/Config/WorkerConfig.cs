namespace BotWorker.Config
{
    public class WorkerConfig
    {
        public string DbConnection { get; set; } = string.Empty;
        public string PluginDirectory { get; set; } = "plugins";
        public string LogLevel { get; set; } = "Information";
        
        public RedisConfig Redis { get; set; } = new();
        public McpConfig Mcp { get; set; } = new();
    }

    public class RedisConfig
    {
        public string Host { get; set; } = "localhost";
        public int Port { get; set; } = 6379;
        public string Password { get; set; } = string.Empty;
    }

    public class McpConfig
    {
        public bool Enabled { get; set; } = true;
        public List<string> AllowedHosts { get; set; } = new();
    }
}


