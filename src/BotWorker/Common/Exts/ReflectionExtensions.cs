using System.Reflection;

namespace BotWorker.Common.Exts
{
    public static class ReflectionExtensions
    {
        // 获取某类型所有公共属性名
        public static IEnumerable<string> GetPublicPropertyNames(this Type type)
            => type.GetProperties(BindingFlags.Public | BindingFlags.Instance).Select(p => p.Name);

        // 根据属性名设置属性值（支持非公开属性）
        public static void SetPropertyValue(this object obj, string propertyName, object? value)
        {
            if (obj == null) throw new ArgumentNullException(nameof(obj));
            var prop = obj.GetType().GetProperty(propertyName,
                BindingFlags.Public | BindingFlags.NonPublic | BindingFlags.Instance);
            if (prop == null || !prop.CanWrite)
                throw new ArgumentException($"Property '{propertyName}' not found or not writable.");
            prop.SetValue(obj, value);
        }

        // 根据属性名获取属性值（支持非公开属性）
        public static object? GetPropertyValue(this object obj, string propertyName)
        {
            if (obj == null) throw new ArgumentNullException(nameof(obj));
            var prop = obj.GetType().GetProperty(propertyName,
                BindingFlags.Public | BindingFlags.NonPublic | BindingFlags.Instance);
            if (prop == null || !prop.CanRead)
                throw new ArgumentException($"Property '{propertyName}' not found or not readable.");
            return prop.GetValue(obj);
        }

        // 判断某类型是否实现了某接口（泛型版本）
        public static bool ImplementsInterface<TInterface>(this Type type)
            => typeof(TInterface).IsAssignableFrom(type);

        // 复制对象属性（属性名相同且可写）
        public static void CopyPropertiesFrom<T>(this T target, T source)
        {
            if (target == null) throw new ArgumentNullException(nameof(target));
            if (source == null) throw new ArgumentNullException(nameof(source));
            var props = typeof(T).GetProperties(BindingFlags.Public | BindingFlags.Instance)
                .Where(p => p.CanRead && p.CanWrite);
            foreach (var prop in props)
            {
                var val = prop.GetValue(source);
                prop.SetValue(target, val);
            }
        }
    }
}
