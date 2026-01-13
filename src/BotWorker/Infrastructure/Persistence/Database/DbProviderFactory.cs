using System.Data;
using System.Data.Common;
using Microsoft.Data.SqlClient;
using Npgsql;
using BotWorker.Common;

namespace BotWorker.Infrastructure.Persistence.Database
{
    public static class DbProviderFactory
    {
        private static readonly System.Threading.AsyncLocal<DatabaseType?> _contextDbType = new();
        private static readonly System.Threading.AsyncLocal<string?> _contextConnString = new();

        public static void SetContext(string connectionString, DatabaseType dbType)
        {
            _contextDbType.Value = dbType;
            _contextConnString.Value = connectionString;
        }

        public static void ClearContext()
        {
            _contextDbType.Value = null;
            _contextConnString.Value = null;
        }

        public static DatabaseType CurrentDbType => _contextDbType.Value ?? GlobalConfig.DbType;
        public static string CurrentConnString => _contextConnString.Value ?? GlobalConfig.ConnString;

        public static IDbConnection CreateConnection()
        {
            return CreateConnection(CurrentConnString, CurrentDbType);
        }

        public static IDbConnection CreateConnection(string connectionString, DatabaseType dbType)
        {
            return dbType switch
            {
                DatabaseType.SqlServer => new SqlConnection(connectionString),
                DatabaseType.PostgreSql => new NpgsqlConnection(connectionString),
                _ => throw new NotSupportedException($"Unsupported database type: {dbType}")
            };
        }

        public static IDataParameter CreateParameter(string name, object? value)
        {
            return CreateParameter(name, value, CurrentDbType);
        }

        public static IDataParameter CreateParameter(string name, object? value, DatabaseType dbType)
        {
            return dbType switch
            {
                DatabaseType.SqlServer => new SqlParameter(FormatParameterName(name), value ?? DBNull.Value),
                DatabaseType.PostgreSql => new NpgsqlParameter(FormatParameterName(name), value ?? DBNull.Value),
                _ => throw new NotSupportedException($"Unsupported database type: {dbType}")
            };
        }

        public static DbDataAdapter CreateDataAdapter(IDbCommand command)
        {
            return CreateDataAdapter(command, CurrentDbType);
        }

        public static DbDataAdapter CreateDataAdapter(IDbCommand command, DatabaseType dbType)
        {
            return dbType switch
            {
                DatabaseType.SqlServer => new SqlDataAdapter((SqlCommand)command),
                DatabaseType.PostgreSql => new NpgsqlDataAdapter((NpgsqlCommand)command),
                _ => throw new NotSupportedException($"Unsupported database type: {dbType}")
            };
        }

        public static string FormatParameterName(string name)
        {
            if (string.IsNullOrEmpty(name)) return name;
            
            // Normalize parameter name prefix
            if (name.StartsWith('@') || name.StartsWith(':') || name.StartsWith('?'))
            {
                name = name[1..];
            }

            return "@" + name;
        }
    }
}
