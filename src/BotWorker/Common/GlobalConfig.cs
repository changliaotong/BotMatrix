namespace BotWorker.Common
{
    public enum DatabaseType
    {
        SqlServer,
        PostgreSql
    }

    public static class GlobalConfig
    {
        private static IConfiguration? _configuration;

        public static IServiceProvider? ServiceProvider { get; set; }

        public static JwtSettings Jwt { get; private set; } = new();
        public static string ConnString { get; set; } = string.Empty;
        public static DatabaseType DbType { get; set; } = DatabaseType.SqlServer;
        public static string RedisConnection { get; set; } = string.Empty;
        public static string SignalRConnString { get; set; } = string.Empty;
        public static string KnowledgeBaseConnection { get; set; } = string.Empty;
        public static string BaseInfoConnection { get; set; } = string.Empty;

        public static DatabaseType GetDatabaseType(string connectionString)
        {
            if (string.IsNullOrEmpty(connectionString)) return DatabaseType.SqlServer;

            // Simple heuristic to detect database type
            if (connectionString.Contains("Host=", StringComparison.OrdinalIgnoreCase) ||
                connectionString.Contains("Port=", StringComparison.OrdinalIgnoreCase) ||
                connectionString.Contains("Username=", StringComparison.OrdinalIgnoreCase))
            {
                return DatabaseType.PostgreSql;
            }

            return DatabaseType.SqlServer;
        }

        public static void Initialize(IConfiguration config)
        {
            _configuration = config ?? throw new ArgumentNullException(nameof(config));

            Jwt = config.GetSection("JwtSettings").Get<JwtSettings>() ?? new JwtSettings();
            ConnString = config.GetConnectionString("DefaultConnection") ?? string.Empty;
            KnowledgeBaseConnection = config.GetConnectionString("KnowledgeBaseConnection") ?? string.Empty;
            BaseInfoConnection = config.GetConnectionString("BaseInfoConnection") ?? string.Empty;
            
            Console.WriteLine($"[CONFIG INFO] Default DB connection string loaded. Length: {ConnString.Length}");
            Console.WriteLine($"[CONFIG INFO] KnowledgeBase connection string loaded. Length: {KnowledgeBaseConnection.Length}");
            Console.WriteLine($"[CONFIG INFO] BaseInfo connection string loaded. Length: {BaseInfoConnection.Length}");
            
            if (ConnString.Length > 0)
            {
                var builder = new System.Data.Common.DbConnectionStringBuilder { ConnectionString = ConnString };
                if (builder.TryGetValue("Server", out var server))
                {
                    Console.WriteLine($"[CONFIG INFO] Database Server: {server}");
                }
            }
            if (Enum.TryParse<DatabaseType>(config["DatabaseType"], true, out var dbType))
            {
                DbType = dbType;
            }
            RedisConnection = config.GetConnectionString("RedisConnection") ?? string.Empty;
            SignalRConnString = config["SignalR:HubUrl"] ?? string.Empty;
        }
        public static string? Get(string key)
        {
            return _configuration?[key];
        }

        public static T GetSection<T>(string sectionName) where T : new()
        {
            if (_configuration == null)
                throw new InvalidOperationException("GlobalConfig is not initialized.");

            var section = new T();
            _configuration.GetSection(sectionName).Bind(section);
            return section;
        }

        public static IConfigurationSection GetSection(string sectionName)
        {
            if (_configuration == null)
                throw new InvalidOperationException("GlobalConfig is not initialized.");

            return _configuration.GetSection(sectionName);
        }

        public static IConfiguration Configuration => _configuration ?? throw new InvalidOperationException("GlobalConfig is not initialized.");
    }

    public class JwtSettings
    {
        /// <summary>
        /// 密钥，用于签署 JWT token
        /// </summary>
        public string SecretKey { get; set; } = string.Empty;

        /// <summary>
        /// Token 颁发者 (Issuer)
        /// </summary>
        public string Issuer { get; set; } = string.Empty;

        /// <summary>
        /// Token 受众 (Audience)
        /// </summary>
        public string Audience { get; set; } = string.Empty;

        /// <summary>
        /// Token 有效期（分钟）
        /// </summary>
        public int ExpiresInHours { get; set; }
    }
}
