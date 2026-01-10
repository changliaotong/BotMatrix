using System.Reflection;

namespace BotWorker.Common.Extensions
{
    public static class ReflectionExtensions
    {
        // ��ȡĳ�������й���������
        public static IEnumerable<string> GetPublicPropertyNames(this Type type)
            => type.GetProperties(BindingFlags.Public | BindingFlags.Instance).Select(p => p.Name);

        // ������������������ֵ��֧�ַǹ������ԣ�
        public static void SetPropertyValue(this object obj, string propertyName, object? value)
        {
            if (obj == null) throw new ArgumentNullException(nameof(obj));
            var prop = obj.GetType().GetProperty(propertyName,
                BindingFlags.Public | BindingFlags.NonPublic | BindingFlags.Instance);
            if (prop == null || !prop.CanWrite)
                throw new ArgumentException($"Property '{propertyName}' not found or not writable.");
            prop.SetValue(obj, value);
        }

        // ������������ȡ����ֵ��֧�ַǹ������ԣ�
        public static object? GetPropertyValue(this object obj, string propertyName)
        {
            if (obj == null) throw new ArgumentNullException(nameof(obj));
            var prop = obj.GetType().GetProperty(propertyName,
                BindingFlags.Public | BindingFlags.NonPublic | BindingFlags.Instance);
            if (prop == null || !prop.CanRead)
                throw new ArgumentException($"Property '{propertyName}' not found or not readable.");
            return prop.GetValue(obj);
        }

        // �ж�ĳ�����Ƿ�ʵ����ĳ�ӿڣ����Ͱ汾��
        public static bool ImplementsInterface<TInterface>(this Type type)
            => typeof(TInterface).IsAssignableFrom(type);

        // ���ƶ������ԣ���������ͬ�ҿ�д��
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


