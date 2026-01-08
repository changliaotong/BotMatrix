namespace BotWorker.Common.Extensions
{
    public static class TypeExtensions
    {
        // 判断对象是否可转换为数字
        public static bool IsNumeric(this object value)
        {
            return double.TryParse(value?.ToString(), out _);
        }

        // ����ǿ��ת��
        public static T? TryCast<T>(this object? obj)
        {
            try { return (T?)Convert.ChangeType(obj, typeof(T)); }
            catch { return default; }
        }
    }
}


