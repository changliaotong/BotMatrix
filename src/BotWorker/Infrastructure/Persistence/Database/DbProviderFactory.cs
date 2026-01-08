using System.Data;
using System.Data.Common;
using Microsoft.Data.SqlClient;
using Npgsql;
using BotWorker.Common;

namespace BotWorker.Infrastructure.Persistence.Database
{
    public static class DbProviderFactory
    {
        public static IDbConnection CreateConnection()
        {
            return GlobalConfig.DbType switch
            {
                DatabaseType.SqlServer => new SqlConnection(GlobalConfig.ConnString),
                DatabaseType.PostgreSql => new NpgsqlConnection(GlobalConfig.ConnString),
                _ => throw new NotSupportedException($"Unsupported database type: {GlobalConfig.DbType}")
            };
        }

        public static IDataParameter CreateParameter(string name, object? value)
        {
            return GlobalConfig.DbType switch
            {
                DatabaseType.SqlServer => new SqlParameter(FormatParameterName(name), value ?? DBNull.Value),
                DatabaseType.PostgreSql => new NpgsqlParameter(FormatParameterName(name), value ?? DBNull.Value),
                _ => throw new NotSupportedException($"Unsupported database type: {GlobalConfig.DbType}")
            };
        }

        public static DbDataAdapter CreateDataAdapter(IDbCommand command)
        {
            return GlobalConfig.DbType switch
            {
                DatabaseType.SqlServer => new SqlDataAdapter((SqlCommand)command),
                DatabaseType.PostgreSql => new NpgsqlDataAdapter((NpgsqlCommand)command),
                _ => throw new NotSupportedException($"Unsupported database type: {GlobalConfig.DbType}")
            };
        }

        public static string FormatParameterName(string name)
        {
            if (string.IsNullOrEmpty(name)) return name;
            
            // Normalize parameter name prefix
            if (name.StartsWith("@") || name.StartsWith(":") || name.StartsWith("?"))
            {
                name = name.Substring(1);
            }

            return "@" + name;
        }
    }
}
