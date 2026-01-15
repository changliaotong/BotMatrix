using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;
using BotWorker.Domain.Repositories;
using BotWorker.Common.Config;
using Dapper;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public class BotCmdRepository : BaseRepository<BotCmd>, IBotCmdRepository
    {
        public BotCmdRepository(string? connectionString = null)
            : base("Cmd", connectionString ?? GlobalConfig.DbConnection)
        {
        }

        public async Task<IEnumerable<string>> GetAllCommandNamesAsync()
        {
            // Original: SELECT CmdText FROM Cmd WHERE IsClose = 0 ORDER BY LENGTH(CmdText) DESC
            // Note: Postgres uses LENGTH(), SqlServer uses LEN(). Dapper should handle or we use generic.
            // But here we can just get all and sort in memory if list is small, or use specific SQL.
            // Assuming Postgres for now as per Program.cs (PostgresDbInitializer).
            string sql = "SELECT \"CmdText\" FROM \"Cmd\" WHERE \"IsClose\" = 0 ORDER BY LENGTH(\"CmdText\") DESC";
            try 
            {
                return await Connection.QueryAsync<string>(sql);
            }
            catch
            {
                // Fallback for SqlServer or if Quotes needed differently
                sql = "SELECT CmdText FROM Cmd WHERE IsClose = 0 ORDER BY LEN(CmdText) DESC";
                 return await Connection.QueryAsync<string>(sql);
            }
        }

        public async Task<string?> GetCmdNameAsync(string cmdText)
        {
            // Original: SELECT CmdName FROM Cmd WHERE CmdText = {cmdText} OR CmdText LIKE '%|{cmdText}|%' ...
            // We use parameters to avoid SQL injection.
            string sql = @"
                SELECT ""CmdName"" FROM ""Cmd"" 
                WHERE ""CmdText"" = @cmdText 
                   OR ""CmdText"" LIKE @pattern1 
                   OR ""CmdText"" LIKE @pattern2 
                   OR ""CmdText"" LIKE @pattern3";
            
            var p = new DynamicParameters();
            p.Add("cmdText", cmdText);
            p.Add("pattern1", $"%|{cmdText}|%");
            p.Add("pattern2", $"{cmdText}|%");
            p.Add("pattern3", $"%|{cmdText}");

            return await Connection.QueryFirstOrDefaultAsync<string>(sql, p);
        }

        public async Task<IEnumerable<string>> GetClosedCommandsAsync()
        {
            // Original: SELECT CmdName FROM Cmd WHERE IsClose = 1
            return await Connection.QueryAsync<string>("SELECT \"CmdName\" FROM \"Cmd\" WHERE \"IsClose\" = 1");
        }

        public async Task<bool> IsCmdCloseAllAsync(string cmdName)
        {
            // Original: GetWhere("IsClose", $"CmdName = {cmdName.Quotes()}").AsBool()
            // Checks if IsClose is true (1) for the given CmdName.
            // Wait, GetWhere returns object, AsBool converts.
            // If result is 1, returns true.
            string sql = "SELECT \"IsClose\" FROM \"Cmd\" WHERE \"CmdName\" = @cmdName";
            var result = await Connection.QueryFirstOrDefaultAsync<int?>(sql, new { cmdName });
            return result.HasValue && result.Value != 0;
        }

        public async Task<string> GetCmdTextAsync(string cmdName)
        {
            string sql = "SELECT \"CmdText\" FROM \"Cmd\" WHERE \"CmdName\" = @cmdName";
            return await Connection.QueryFirstOrDefaultAsync<string>(sql, new { cmdName }) ?? string.Empty;
        }

        public async Task EnsureCommandExistsAsync(string name, string text)
        {
            string checkSql = "SELECT COUNT(*) FROM \"Cmd\" WHERE \"CmdName\" = @name";
            var count = await Connection.ExecuteScalarAsync<int>(checkSql, new { name });
            
            if (count == 0)
            {
                var cmd = new BotCmd
                {
                    CmdName = name,
                    CmdText = text,
                    IsClose = 0
                };
                await AddAsync(cmd);
            }
        }
    }
}
