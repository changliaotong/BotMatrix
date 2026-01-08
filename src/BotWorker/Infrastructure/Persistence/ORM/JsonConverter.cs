using System.Reflection;
using System.Text.Json;

namespace BotWorker.Infrastructure.Persistence.ORM
{
    public class JsonConverter : IValueConverter
    {
        public object? ConvertToProvider(object? value)
        {
            if (value == null) return DBNull.Value;
            return JsonSerializer.Serialize(value);
        }

        public object? ConvertFromProvider(object? value)
        {
            if (value == null || value == DBNull.Value) return null;
            return JsonSerializer.Deserialize<object>(value.ToString()!);
        }

        /// <summary>
        /// 获取转换后的列值，自动应用 ConverterType 转换器（如果有）
        /// </summary>
        public static object? GetColumnValue(object entity, PropertyInfo prop)
        {
            var value = prop.GetValue(entity);
            var attr = prop.GetCustomAttribute<ColumnAttribute>();
            if (attr?.ConverterType != null)
            {
                if (Activator.CreateInstance((Type)attr.ConverterType) is IValueConverter converter)
                {
                    return converter.ConvertToProvider(value) ?? DBNull.Value;
                }
                else
                {
                    throw new InvalidOperationException($"ConverterType {attr.ConverterType} must implement IValueConverter");
                }
            }
            return value ?? DBNull.Value;
        }
    }

}
