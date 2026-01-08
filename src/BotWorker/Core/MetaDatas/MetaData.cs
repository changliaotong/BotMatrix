using System.Data;
using System.Reflection;
using Microsoft.Data.SqlClient;
using Microsoft.EntityFrameworkCore.Metadata.Internal;
using Newtonsoft.Json;
using BotWorker.Core.Database.Mapping;

namespace BotWorker.Core.MetaDatas
{
    public abstract partial class MetaData<TDerived> where TDerived : MetaData<TDerived>, new()
    {
        [JsonIgnore]
        public virtual string DataBase { get; } = "sz84_robot";
        [JsonIgnore]
        public virtual string TableName { get; }
        [JsonIgnore]
        public virtual TimeSpan CacheTime { get; } = TimeSpan.FromMinutes(30);
        [JsonIgnore]
        public abstract string KeyField { get; }
        [JsonIgnore]
        public virtual string KeyField2 { get; } = string.Empty;
        [JsonIgnore]
        public static string Key { get; set; } = string.Empty;
        [JsonIgnore]
        public static string Key2 { get; set; } = string.Empty;
        [JsonIgnore]
        public static string DbName { get; set; } = string.Empty;
        [JsonIgnore]
        public virtual IReadOnlyList<string> KeyFields =>
            string.IsNullOrEmpty(KeyField2) ? [KeyField] : new[] { KeyField, KeyField2 };

        // 静态缓存子类的主键信息和完整表名
        [JsonIgnore]
        public static readonly string[] Keys;
        [JsonIgnore]
        public static readonly string FullName;

        // 实例方法访问静态缓存字段，方便服务层通过实例获取信息        
        public IReadOnlyList<string> GetKeys() => Keys;        
        public string GetFullName() => FullName;
        private static readonly TDerived _instance = new();

        static MetaData()
        {
            var instance = _instance;
            DbName = instance.DataBase;
            Keys = [instance.KeyField, instance.KeyField2];
            Key = instance.KeyField;
            Key2 = instance.KeyField2;
            FullName = $"[{instance.DataBase}].[dbo].[{instance.TableName}]";
        }

        // 查询：直接静态调用，内部用单例实例处理
        public static Task<List<TDerived>> QueryListAsync(QueryOptions? options = null)
            => _instance.GetListAsync(options);

        public Dictionary<string, object?> ToDictionary()
        {
            var dict = new Dictionary<string, object?>();
            var props = GetType().GetProperties();

            foreach (var prop in props)
            {
                // 1. 跳过索引器属性（带参数的属性，不能直接获取）
                if (prop.GetIndexParameters().Length > 0)
                    continue;

                // 2. 跳过标记了 [DbIgnore] 的属性（自定义显式排除）
                if (prop.GetCustomAttribute<DbIgnoreAttribute>() != null)
                    continue;

                // 3. 跳过只读属性（没有 setter，通常是计算属性，不存数据库）
                if (!prop.CanWrite)
                    continue;

                // 4. 跳过非公共读写属性（一般不存数据库）
                if (!prop.CanRead || !prop.GetMethod!.IsPublic || !prop.SetMethod!.IsPublic)
                    continue;

                // 5. 可选：跳过静态属性（静态字段不属于实例，不存数据库）
                if (prop.GetMethod!.IsStatic)
                    continue;

                // 6. 可选：跳过索引器或特殊属性名，比如以 "_" 或 "$" 开头的（业务需求）
                if (prop.Name.StartsWith("_") || prop.Name.StartsWith("$"))
                    continue;

                // 7. 这里可加你业务特殊判断，比如排除某些字段名等

                var value = prop.GetValue(this);

                dict[prop.Name] = value;
            }

            return dict;
        }

        public static async Task<TDerived> LoadAsync(object id, object? id2 = null)
        {
            return await GetSingleAsync(id, id2) ?? throw new Exception($"主键属性 {id} {id2}不存在");
        }

        // 返回主键列表，保持顺序，方便生成SQL和参数绑定
        public List<(string Name, object Value)> GetKeyValues()
        {
            var list = new List<(string, object)>();
            foreach (var key in Keys)
            {
                var prop = typeof(TDerived).GetProperty(key) ?? throw new Exception($"主键属性 {key} 不存在");
                list.Add((key, prop.GetValue(this) ?? DBNull.Value));
            }
            return list;
        }

        protected virtual Dictionary<string, object> GetInsertFields()
        {
            return PropertyHelper.GetAll(GetType())
                        .Where(p => p.IncludeInInsert)
                        .ToDictionary(p => p.ColumnName, p => p.GetValue(this) ?? DBNull.Value);
        }

        protected virtual Dictionary<string, object> GetUpdateFields()
        {
            return PropertyHelper.GetAll(GetType())
                .Where(p => p.IncludeInUpdate)
                .ToDictionary(p => p.ColumnName, p => p.GetValue(this) ?? DBNull.Value);
        }

        public static string GetSqlValue(object value, string parameterName)
        {
            if (value is DateTime dateTimeValue && dateTimeValue == DateTime.MinValue)
            {
                return $"ISNULL({parameterName}, GETDATE())";
            }
            else if (value is Guid guidValue && guidValue == Guid.Empty)
            {
                return $"ISNULL({parameterName}, NEWID())";
            }
            else
            {
                return parameterName;
            }
        }

        public static SqlParameter GetSqlParameter(string parameterName, object value)
        {
            if (value is byte[] byteValue)
            {
                return new SqlParameter(parameterName, SqlDbType.VarBinary) { Value = byteValue };
            }
            else if (value is bool boolValue)
            {
                return new SqlParameter(parameterName, boolValue ? 1 : 0);
            }
            else if (value is Enum enumValue)
            {
                return new SqlParameter(parameterName, Convert.ToInt32(enumValue));
            }
            else if (value is string strValue)
            {
                return new SqlParameter(parameterName, string.IsNullOrEmpty(strValue) ? "" : strValue);
            }
            else
            {
                return new SqlParameter(parameterName, value ?? DBNull.Value);
            }
        } 

        public static string FormatValue(object value)
        {
            if (value is null)
            {
                return "NULL";
            }
            else if (value is string)
            {
                return $"N'{EscapeSqlString(value.AsString())}'";
            }
            else if (value is DateTime dateTime)
            {
                if (dateTime == DateTime.MinValue)
                {
                    return "GETDATE()";
                }
                else
                {
                    return $"'{dateTime:yyyy-MM-dd HH:mm:ss}'";
                }
            }
            else
            {
                return value.AsString();
            }
        }

        private static string EscapeSqlString(string value)
        {
            // 在需要的情况下对字符串中的特殊字符进行转义，以防止SQL注入攻击
            // 这里只是简单地对单引号进行替换，更复杂的情况可能需要更多的处理
            return value.Replace("'", "''");
        }
    }
}
