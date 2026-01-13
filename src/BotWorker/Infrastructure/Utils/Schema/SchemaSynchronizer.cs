using System.Reflection;
using System.Text;
using BotWorker.Infrastructure.Utils.Schema.Attributes;

namespace BotWorker.Infrastructure.Utils.Schema;

public static class SchemaSynchronizer
{
    public static string GenerateCreateTableSql<T>() where T : class
    {
        var type = typeof(T);
        
        // 尝试获取 FullName (针对 MetaData 子类)
        string? fullName = null;
        var fullNameField = type.GetField("FullName", BindingFlags.Public | BindingFlags.Static | BindingFlags.FlattenHierarchy);
        if (fullNameField != null)
        {
            fullName = fullNameField.GetValue(null)?.ToString();
        }

        var tableName = fullName ?? $"[{TableNameHelper.GetTableName<T>()}]";

        var sb = new StringBuilder();
        sb.AppendLine($"CREATE TABLE {tableName} (");

        var properties = type.GetProperties(BindingFlags.Public | BindingFlags.Instance);
        var primaryKeyColumn = "";

        foreach (var prop in properties)
        {
            if (prop.GetCustomAttribute<IgnoreColumnAttribute>() != null)
                continue;

            var colAttr = prop.GetCustomAttribute<BotWorker.Infrastructure.Utils.Schema.Attributes.ColumnAttribute>();
            var columnName = colAttr?.Name ?? prop.Name;
            
            bool isPrimaryKey = prop.Name.Equals("Id", StringComparison.OrdinalIgnoreCase) ||
                                prop.GetCustomAttribute<PrimaryKeyAttribute>() != null;

            var sqlType = SqlTypeMapper.Map(prop.PropertyType, isPrimaryKey);

            sb.Append($"  [{columnName}] {sqlType}");

            // 主键
            if (isPrimaryKey)
            {
                // 如果是 SQL Server 且是数字类型，添加 IDENTITY
                var isSqlServer = !BotWorker.Infrastructure.Persistence.ORM.MetaData.IsPostgreSql;
                var propType = Nullable.GetUnderlyingType(prop.PropertyType) ?? prop.PropertyType;
                if (isSqlServer && (propType == typeof(int) || propType == typeof(long)))
                {
                    sb.Append(" IDENTITY(1,1)");
                }
                else if (isSqlServer && propType == typeof(Guid))
                {
                    sb.Append(" DEFAULT NEWID()");
                }

                sb.Append(" PRIMARY KEY");
                primaryKeyColumn = columnName;
            }

            sb.AppendLine(",");
        }

        sb.Length -= 3; // 去掉最后逗号
        sb.AppendLine("\n);");

        return sb.ToString();
    }
}
