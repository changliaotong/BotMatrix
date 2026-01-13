using System;
using System.IO;
using System.Threading.Tasks;
using BotWorker.Common;
using Dapper;
using Microsoft.Extensions.Logging;
using Npgsql;

namespace BotWorker.Modules.AI.Services
{
    public interface IDbInitializer
    {
        Task InitializeAsync();
    }

    public class PostgresDbInitializer : IDbInitializer
    {
        private readonly string _connectionString;
        private readonly ILogger<PostgresDbInitializer> _logger;

        public PostgresDbInitializer(ILogger<PostgresDbInitializer> logger)
        {
            _connectionString = GlobalConfig.KnowledgeBaseConnection;
            _logger = logger;
        }

        public async Task InitializeAsync()
        {
            try
            {
                _logger.LogInformation("Starting AI database initialization...");

                using var conn = new NpgsqlConnection(_connectionString);
                await conn.OpenAsync();

                // 读取初始化脚本
                string scriptPath = Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "scripts", "db", "init_ai_pg.sql");
                
                // 考虑到开发环境下路径可能不同，尝试几个可能的位置
                if (!File.Exists(scriptPath))
                {
                    // 尝试项目根目录下的 scripts
                    scriptPath = Path.Combine(Directory.GetCurrentDirectory(), "scripts", "db", "init_ai_pg.sql");
                }
                
                if (!File.Exists(scriptPath))
                {
                    // 再次尝试
                    scriptPath = "scripts/db/init_ai_pg.sql";
                }

                if (!File.Exists(scriptPath))
                {
                    _logger.LogError("Could not find database initialization script at {Path}", scriptPath);
                    return;
                }

                _logger.LogInformation("Executing database initialization script from {Path}", scriptPath);
                string sql = await File.ReadAllTextAsync(scriptPath);

                // 执行 SQL
                await conn.ExecuteAsync(sql);

                _logger.LogInformation("AI database initialization completed successfully.");
            }
            catch (Exception ex)
            {
                _logger.LogError(ex, "Failed to initialize AI database.");
                throw;
            }
        }
    }
}
