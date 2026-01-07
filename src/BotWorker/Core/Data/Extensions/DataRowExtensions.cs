using System.Data;
using System.Reflection;
using System.Text.Json;

namespace BotWorker.Core.Data.Extensions;

public static class DataRowExtensions
{
    public static Dictionary<string, object> ToDictionary(this DataRow row)
    {
        var dict = new Dictionary<string, object>(StringComparer.OrdinalIgnoreCase);
        foreach (DataColumn col in row.Table.Columns)
        {
            var value = row[col];
            dict[col.ColumnName] = value == DBNull.Value ? null! : value;
        }
        return dict;
    }

    public static T ToModel<T>(this DataRow row) where T : new()
    {
        var model = new T();
        var type = typeof(T);
        var properties = type.GetProperties(BindingFlags.Public | BindingFlags.Instance);

        foreach (var prop in properties)
        {
            if (!row.Table.Columns.Contains(prop.Name)) continue;

            var value = row[prop.Name];

            if (value == DBNull.Value)
            {
                // 自动填充时间字段
                if (prop.PropertyType == typeof(DateTime) || prop.PropertyType == typeof(DateTime?))
                {
                    if (prop.Name.Equals("CreateTime", StringComparison.OrdinalIgnoreCase) ||
                        prop.Name.Equals("UpdateTime", StringComparison.OrdinalIgnoreCase))
                    {
                        prop.SetValue(model, DateTime.Now);
                    }
                }
                continue;
            }

            try
            {
                // JSON 字段自动反序列化（适用于 List<string>、List<int> 等）
                if (prop.PropertyType != typeof(string) &&
                    prop.PropertyType != typeof(DateTime) &&
                    prop.PropertyType != typeof(DateTime?) &&
                    prop.PropertyType.IsGenericType &&
                    (prop.PropertyType.GetGenericTypeDefinition() == typeof(List<>)))
                {
                    var json = value.ToString();
                    if (!string.IsNullOrWhiteSpace(json))
                    {
                        var deserialized = JsonSerializer.Deserialize(json, prop.PropertyType);
                        if (deserialized != null)
                            prop.SetValue(model, deserialized);
                        continue;
                    }
                }

                // 普通转换
                var safeValue = Convert.ChangeType(value, Nullable.GetUnderlyingType(prop.PropertyType) ?? prop.PropertyType);
                prop.SetValue(model, safeValue);
            }
            catch
            {
                // 忽略字段赋值错误
            }
        }

        return model;
    }
}
