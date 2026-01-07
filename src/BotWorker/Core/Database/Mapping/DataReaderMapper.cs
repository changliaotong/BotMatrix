using System.Collections.Concurrent;
using Microsoft.Data.SqlClient;
using System.Reflection;

namespace sz84.Core.Database.Mapping
{
    public static class DataReaderMapper
    {
        // 缓存实体属性信息，避免反射重复获取
        private static readonly ConcurrentDictionary<Type, PropertyInfo[]> PropertyCache = new();

        public static T MapToEntity<T>(SqlDataReader reader) where T : new()
        {
            var type = typeof(T);
            var props = PropertyCache.GetOrAdd(type, t => t.GetProperties(BindingFlags.Public | BindingFlags.Instance));

            var entity = new T();

            for (int i = 0; i < reader.FieldCount; i++)
            {
                var name = reader.GetName(i);
                var prop = Array.Find(props, p => string.Equals(p.Name, name, StringComparison.OrdinalIgnoreCase));
                if (prop != null && !reader.IsDBNull(i))
                {
                    var val = reader.GetValue(i);
                    try
                    {
                        // 兼容Nullable和基础类型
                        var targetType = Nullable.GetUnderlyingType(prop.PropertyType) ?? prop.PropertyType;
                        var safeValue = Convert.ChangeType(val, targetType);
                        prop.SetValue(entity, safeValue);
                    }
                    catch
                    {
                        // 可日志警告，或忽略赋值失败，防止整个映射失败
                    }
                }
            }

            return entity;
        }
    }

}
