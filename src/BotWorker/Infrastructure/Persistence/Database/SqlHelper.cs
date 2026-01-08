using System.Reflection;

namespace BotWorker.Infrastructure.Persistence.Database
{
    public static class SqlHelper
    {
        public static T ConvertValue<T>(object? value, T? defaultValue = default)
        {
            if (value == null || value == DBNull.Value)
                return defaultValue!;

            var targetType = typeof(T);

            try
            {
                // Guid 特殊处理
                if (targetType == typeof(Guid))
                {
                    if (value is Guid g)
                        return (T)(object)g;

                    if (value is byte[] bytes && bytes.Length == 16)
                        return (T)(object)new Guid(bytes);

                    if (Guid.TryParse(value.ToString(), out var guid))
                        return (T)(object)guid;

                    return defaultValue!;
                }

                // Enum 特殊处理
                if (targetType.IsEnum)
                {
                    return (T)Enum.Parse(targetType, value.ToString()!, ignoreCase: true);
                }

                // bool 特殊处理
                if (targetType == typeof(bool))
                {
                    if (value is int i)
                        return (T)(object)(i != 0);
                    if (value is string s)
                        return (T)(object)(s == "1" || s.Equals("true", StringComparison.OrdinalIgnoreCase));
                    if (value is bool b)
                        return (T)(object)b;
                }

                // 其它类型正常转换
                return (T)Convert.ChangeType(value, targetType);
            }
            catch
            {
                return defaultValue!;
            }
        }
        

        public static string GetColumnName(PropertyInfo prop)
        {
            // 先尝试获取自定义的 ColumnAttribute 特性
            var attr = prop.GetCustomAttribute<ColumnAttribute>();
            if (attr != null && !string.IsNullOrEmpty(attr.Name))
            {
                return attr.Name;
            }
            // 没有特性或没指定名称就用属性名
            return prop.Name;
        }

        public static object? ConvertFromDbValue(object? dbValue, PropertyInfo prop)
        {
            var attr = prop.GetCustomAttribute<ColumnAttribute>();
            if (attr?.ConverterType != null)
            {
                if (Activator.CreateInstance((Type)attr.ConverterType) is IValueConverter converter)
                {
                    return converter.ConvertFromProvider(dbValue);
                }
                else
                {
                    throw new InvalidOperationException($"ConverterType {attr.ConverterType} must implement IValueConverter");
                }
            }

            if (dbValue == DBNull.Value) return null;
            return dbValue;
        }

    }
}

