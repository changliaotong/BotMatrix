using System;
using System.Collections.Concurrent;
using System.Collections.Generic;
using System.Data;
using System.Linq;
using System.Reflection;
using System.Threading.Tasks;
using BotWorker.Common;
using BotWorker.Domain.Repositories;
using BotWorker.Infrastructure.Persistence;
using Dapper;
using Dapper.Contrib.Extensions;
using Npgsql;

namespace BotWorker.Infrastructure.Persistence.Repositories
{
    public abstract class BaseRepository<T> : IBaseRepository<T> where T : class
    {
        protected readonly string _connectionString;
        protected readonly string _tableName;
        protected virtual string KeyField => "id";
        
        private static readonly ConcurrentDictionary<Type, List<PropertyInfo>> _paramCache = new();

        protected BaseRepository(string tableName, string? connectionString = null)
        {
            _tableName = tableName;
            _connectionString = connectionString ?? GlobalConfig.KnowledgeBaseConnection;
            
            // 确保 Dapper 支持下划线命名映射到帕斯卡命名
            DefaultTypeMap.MatchNamesWithUnderscores = true;
        }

        protected IDbConnection CreateConnection() => new NpgsqlConnection(_connectionString);

        private bool _isTableChecked = false;
        public virtual async Task EnsureTableCreatedAsync()
        {
            if (_isTableChecked) return;
            try
            {
                using var conn = CreateConnection();
                // 简单的 PGSQL 检查，如果连接字符串包含 Host= 则认为是 PG
                string sqlCheck = _connectionString.Contains("Host=")
                    ? $"SELECT COUNT(*) FROM pg_tables WHERE schemaname = 'public' AND tablename = '{_tableName.ToLower()}'"
                    : $"SELECT COUNT(*) FROM sys.tables WHERE name = '{_tableName}'";

                var count = await conn.ExecuteScalarAsync<int>(sqlCheck);
                if (count == 0)
                {
                    var sqlCreate = BotWorker.Infrastructure.Utils.Schema.SchemaSynchronizer.GenerateCreateTableSql<T>();
                    await conn.ExecuteAsync(sqlCreate);
                }
                _isTableChecked = true;
            }
            catch (Exception ex)
            {
                Console.WriteLine($"[ORM] Error ensuring table {_tableName} exists: {ex.Message}");
            }
        }

        public virtual async Task<SqlHelper.TransactionWrapper> BeginTransactionAsync(IDbTransaction? existingTrans = null)
        {
            return await SqlHelper.BeginTransactionAsync(existingTrans);
        }

        public virtual async Task<T?> GetByIdAsync(long id)
        {
            using var conn = CreateConnection();
            return await conn.GetAsync<T>(id);
        }

        public virtual async Task<IEnumerable<T>> GetAllAsync()
        {
            using var conn = CreateConnection();
            return await conn.GetAllAsync<T>();
        }

        public virtual async Task<bool> DeleteAsync(long id)
        {
            using var conn = CreateConnection();
            return await conn.ExecuteAsync($"DELETE FROM {_tableName} WHERE {KeyField} = @id", new { id }) > 0;
        }

        public virtual async Task<IEnumerable<T>> GetListAsync(string? conditions = null, object? parameters = null, IDbTransaction? trans = null)
        {
            string sql = $"SELECT * FROM {_tableName}";
            if (!string.IsNullOrEmpty(conditions))
            {
                sql += " " + conditions;
            }
            if (trans != null)
            {
                return await trans.Connection.QueryAsync<T>(sql, parameters, trans);
            }
            using var conn = CreateConnection();
            return await conn.QueryAsync<T>(sql, parameters);
        }

        public virtual async Task<T?> GetFirstOrDefaultAsync(string? conditions = null, object? parameters = null, IDbTransaction? trans = null)
        {
            string sql = $"SELECT * FROM {_tableName}";
            if (!string.IsNullOrEmpty(conditions))
            {
                sql += " " + conditions;
            }
            if (trans != null)
            {
                return await trans.Connection.QueryFirstOrDefaultAsync<T>(sql, parameters, trans);
            }
            using var conn = CreateConnection();
            return await conn.QueryFirstOrDefaultAsync<T>(sql, parameters);
        }

        public virtual async Task<long> CountAsync(string? conditions = null, object? parameters = null, IDbTransaction? trans = null)
        {
            string sql = $"SELECT COUNT(1) FROM {_tableName}";
            if (!string.IsNullOrEmpty(conditions))
            {
                sql += " " + conditions;
            }
            if (trans != null)
            {
                return await trans.Connection.ExecuteScalarAsync<long>(sql, parameters, trans);
            }
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<long>(sql, parameters);
        }

        public virtual async Task<long> InsertAsync(T entity, IDbTransaction? trans = null)
        {
            var type = typeof(T);
            var props = GetScannableProperties(type);
            
            // Filter out [Key] (Identity) - usually skipped in insert unless ExplicitKey
            // Dapper.Contrib: [Key] is identity (skip), [ExplicitKey] is manual (include)
            var insertProps = props.Where(p => 
                !p.GetCustomAttributes(typeof(KeyAttribute), true).Any() &&
                !p.GetCustomAttributes(typeof(ComputedAttribute), true).Any()
            ).ToList();

            var columns = insertProps.Select(p => ToSnakeCase(p.Name));
            var paramNames = insertProps.Select(p => "@" + p.Name);
            
            var sql = $"INSERT INTO {_tableName} ({string.Join(", ", columns)}) VALUES ({string.Join(", ", paramNames)})";
            
            // If there is a [Key] property, we might want to return it.
            // Postgres: RETURNING id
            var keyProp = props.FirstOrDefault(p => p.GetCustomAttributes(typeof(KeyAttribute), true).Any());
            if (keyProp != null)
            {
                 var keyCol = ToSnakeCase(keyProp.Name);
                 sql += $" RETURNING {keyCol}";
                 if (trans != null) return await trans.Connection.ExecuteScalarAsync<long>(sql, entity, trans);
                 using var conn = CreateConnection();
                 return await conn.ExecuteScalarAsync<long>(sql, entity);
            }
            
            if (trans != null)
            {
                await trans.Connection.ExecuteAsync(sql, entity, trans);
                return 0;
            }
            using var conn = CreateConnection();
            await conn.ExecuteAsync(sql, entity);
            return 0;
        }

        public virtual async Task<bool> UpdateEntityAsync(T entity, IDbTransaction? trans = null)
        {
            var type = typeof(T);
            var props = GetScannableProperties(type);
            
            var keyProp = props.FirstOrDefault(p => 
                p.GetCustomAttributes(typeof(KeyAttribute), true).Any() || 
                p.GetCustomAttributes(typeof(ExplicitKeyAttribute), true).Any());
            
            if (keyProp == null)
                keyProp = type.GetProperties().FirstOrDefault(p => p.Name.Equals("Id", StringComparison.OrdinalIgnoreCase));

            if (keyProp == null) throw new ArgumentException("Entity must have a [Key], [ExplicitKey] or Id property");

            var updateProps = props.Where(p => 
                p != keyProp &&
                !p.GetCustomAttributes(typeof(ComputedAttribute), true).Any()
            ).ToList();

            var setClause = string.Join(", ", updateProps.Select(p => $"{ToSnakeCase(p.Name)} = @{p.Name}"));
            var keyName = ToSnakeCase(keyProp.Name);
            
            var sql = $"UPDATE {_tableName} SET {setClause} WHERE {keyName} = @{keyProp.Name}";

            if (trans != null) return await trans.Connection.ExecuteAsync(sql, entity, trans) > 0;
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, entity) > 0;
        }

        private List<PropertyInfo> GetScannableProperties(Type type)
        {
            if (_paramCache.TryGetValue(type, out var props)) return props;
            
            props = type.GetProperties().Where(p => 
                !p.GetCustomAttributes(typeof(WriteAttribute), true).Any(a => !((WriteAttribute)a).Write)
            ).ToList();
            _paramCache[type] = props;
            return props;
        }

        protected string ToSnakeCase(string input)
        {
            if (string.IsNullOrEmpty(input)) return input;
            var startUnderscore = input.StartsWith("_");
            var res = System.Text.RegularExpressions.Regex.Replace(input, @"([a-z0-9])([A-Z])", "$1_$2").ToLower();
            return startUnderscore ? "_" + res : res;
        }

        public virtual async Task<TValue> GetValueAsync<TValue>(string field, object keyValue, string keyField, IDbTransaction? trans = null)
        {
            if (!System.Text.RegularExpressions.Regex.IsMatch(field, @"^[a-zA-Z0-9_]+$"))
                throw new ArgumentException("Invalid field name");
            if (!System.Text.RegularExpressions.Regex.IsMatch(keyField, @"^[a-zA-Z0-9_]+$"))
                throw new ArgumentException("Invalid key field name");

            string dbField = field.Contains("(") ? field : ToSnakeCase(field);
            string dbKeyField = ToSnakeCase(keyField);
            string sql = $"SELECT {dbField} FROM {_tableName} WHERE {dbKeyField} = @keyValue";
            
            if (trans != null)
            {
                return await trans.Connection.ExecuteScalarAsync<TValue>(sql, new { keyValue }, trans);
            }
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<TValue>(sql, new { keyValue });
        }

        public virtual async Task<TValue> GetValueAsync<TValue>(string field, string conditions, object? parameters = null, IDbTransaction? trans = null)
        {
            if (!System.Text.RegularExpressions.Regex.IsMatch(field, @"^[a-zA-Z0-9_\(\)\+\s]+$")) // Allow COUNT(1) + 1 etc.
                throw new ArgumentException("Invalid field name");

            string dbField = field.Contains("(") ? field : ToSnakeCase(field);
            string sql = $"SELECT {dbField} FROM {_tableName} " + conditions;
            if (trans != null)
            {
                return await trans.Connection.ExecuteScalarAsync<TValue>(sql, parameters, trans);
            }
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<TValue>(sql, parameters);
        }

        public virtual async Task<T> ExecuteScalarAsync<T>(string sql, object? parameters = null, IDbTransaction? trans = null)
        {
            if (trans != null)
            {
                return await trans.Connection.ExecuteScalarAsync<T>(sql, parameters, trans);
            }
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<T>(sql, parameters);
        }

        public virtual async Task<int> SetValueAsync(string field, object value, string conditions, object? parameters = null, IDbTransaction? trans = null)
        {
            if (!System.Text.RegularExpressions.Regex.IsMatch(field, @"^[a-zA-Z0-9_]+$"))
                throw new ArgumentException("Invalid field name");

            string dbField = ToSnakeCase(field);
            string sql = $"UPDATE {_tableName} SET {dbField} = @value, updated_at = CURRENT_TIMESTAMP " + conditions;
            
            // Merge value into parameters if possible, but Dapper handles anonymous types.
            // We need to combine 'value' and 'parameters'.
            // The cleanest way is to create a DynamicParameters object.
            var p = new DynamicParameters(parameters);
            p.Add("value", value);

            if (trans != null)
            {
                return await trans.Connection.ExecuteAsync(sql, p, trans);
            }
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, p);
        }

        public virtual async Task<int> IncrementValueAsync(string field, object value, string conditions, object? parameters = null, IDbTransaction? trans = null)
        {
            if (!System.Text.RegularExpressions.Regex.IsMatch(field, @"^[a-zA-Z0-9_]+$"))
                throw new ArgumentException("Invalid field name");

            string dbField = ToSnakeCase(field);
            string sql = $"UPDATE {_tableName} SET {dbField} = {dbField} + @value, updated_at = CURRENT_TIMESTAMP " + conditions;
            
            var p = new DynamicParameters(parameters);
            p.Add("value", value);

            if (trans != null)
            {
                return await trans.Connection.ExecuteAsync(sql, p, trans);
            }
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, p);
        }

        public virtual async Task<TValue> GetValueAsync<TValue>(string field, long id, IDbTransaction? trans = null)
        {
            if (!System.Text.RegularExpressions.Regex.IsMatch(field, @"^[a-zA-Z0-9_]+$"))
                throw new ArgumentException("Invalid field name");

            string dbField = ToSnakeCase(field);
            string sql = $"SELECT {dbField} FROM {_tableName} WHERE {KeyField} = @id";
            if (trans != null)
            {
                return await trans.Connection.ExecuteScalarAsync<TValue>(sql, new { id }, trans);
            }
            using var conn = CreateConnection();
            return await conn.ExecuteScalarAsync<TValue>(sql, new { id });
        }

        public virtual async Task<int> SetValueAsync(string field, object value, long id, IDbTransaction? trans = null)
        {
            if (!System.Text.RegularExpressions.Regex.IsMatch(field, @"^[a-zA-Z0-9_]+$"))
                throw new ArgumentException("Invalid field name");

            string dbField = ToSnakeCase(field);
            string sql = $"UPDATE {_tableName} SET {dbField} = @value, updated_at = CURRENT_TIMESTAMP WHERE {KeyField} = @id";
            
            if (trans != null)
            {
                return await trans.Connection.ExecuteAsync(sql, new { value, id }, trans);
            }
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, new { value, id });
        }

        public virtual async Task<int> IncrementValueAsync(string field, object value, long id, IDbTransaction? trans = null)
        {
            if (!System.Text.RegularExpressions.Regex.IsMatch(field, @"^[a-zA-Z0-9_]+$"))
                throw new ArgumentException("Invalid field name");

            string dbField = ToSnakeCase(field);
            string sql = $"UPDATE {_tableName} SET {dbField} = {dbField} + @value, updated_at = CURRENT_TIMESTAMP WHERE {KeyField} = @id";
            
            if (trans != null)
            {
                return await trans.Connection.ExecuteAsync(sql, new { value, id }, trans);
            }
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, new { value, id });
        }

        public virtual async Task<int> UpdateAsync(string fieldsSql, long id, IDbTransaction? trans = null)
        {
            string sql = $"UPDATE {_tableName} SET {fieldsSql}, updated_at = CURRENT_TIMESTAMP WHERE {KeyField} = @id";
            if (trans != null)
            {
                return await trans.Connection.ExecuteAsync(sql, new { id }, trans);
            }
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, new { id });
        }

        public virtual async Task<int> UpdateAsync(string fieldsSql, string conditions, object? parameters = null, IDbTransaction? trans = null)
        {
            string sql = $"UPDATE {_tableName} SET {fieldsSql}, updated_at = CURRENT_TIMESTAMP {conditions}";
            if (trans != null)
            {
                return await trans.Connection.ExecuteAsync(sql, parameters, trans);
            }
            using var conn = CreateConnection();
            return await conn.ExecuteAsync(sql, parameters);
        }
    }
}
