using System.Reflection;
using System.Text;
using sz84.Infrastructure.Utils.Schema.Attributes;

namespace sz84.Infrastructure.Utils.Schema;

public static class SchemaSynchronizer
{
    public static string GenerateCreateTableSql<T>() where T : class
    {
        var type = typeof(T);
        var tableName = TableNameHelper.GetTableName<T>();

        var sb = new StringBuilder();
        sb.AppendLine($"CREATE TABLE [{tableName}] (");

        var properties = type.GetProperties(BindingFlags.Public | BindingFlags.Instance);
        var primaryKeyColumn = "";

        foreach (var prop in properties)
        {
            if (prop.GetCustomAttribute<IgnoreColumnAttribute>() != null)
                continue;

            var colAttr = prop.GetCustomAttribute<ColumnAttribute>();
            var columnName = colAttr?.Name ?? prop.Name.ToLower();
            var sqlType = SqlTypeMapper.Map(prop.PropertyType);

            sb.Append($"  [{columnName}] {sqlType}");

            // 主键
            if (prop.Name.Equals("Id", StringComparison.OrdinalIgnoreCase) ||
                prop.GetCustomAttribute<PrimaryKeyAttribute>() != null)
            {
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
