using System;
using System.Collections.Generic;
using System.Data;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Common;
using Dapper;
using Npgsql;

namespace BotWorker.Modules.AI.Repositories
{
    public abstract class BasePostgresRepository<T> where T : class
    {
        protected readonly string _connectionString;
        protected readonly string _tableName;

        protected BasePostgresRepository(string tableName, string? connectionString = null)
        {
            _tableName = tableName;
            _connectionString = connectionString ?? GlobalConfig.KnowledgeBaseConnection;
            
            // 确保 Dapper 支持下划线命名映射到帕斯卡命名
            DefaultTypeMap.MatchNamesWithUnderscores = true;
        }

        protected IDbConnection CreateConnection() => new NpgsqlConnection(_connectionString);

        public virtual async Task<T?> GetByIdAsync(long id)
        {
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<T>(
                $"SELECT * FROM {_tableName} WHERE id = @id", new { id });
        }

        public virtual async Task<IEnumerable<T>> GetAllAsync()
        {
            using var conn = CreateConnection();
            return await conn.QueryAsync<T>($"SELECT * FROM {_tableName} ORDER BY id DESC");
        }

        public virtual async Task<bool> DeleteAsync(long id)
        {
            using var conn = CreateConnection();
            return await conn.ExecuteAsync($"DELETE FROM {_tableName} WHERE id = @id", new { id }) > 0;
        }

        // AddAsync 和 UpdateAsync 通常需要特定的 SQL，所以留给子类实现或提供辅助方法
    }
}
