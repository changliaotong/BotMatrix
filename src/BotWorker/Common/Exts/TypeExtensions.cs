namespace BotWorker.Common.Exts
{
    public static class TypeExtensions
    {
        // 判断对象是否可转换为数字
        public static bool IsNumeric(this object value)
        {
            return double.TryParse(value?.ToString(), out _);
        }

        // 尝试强制转换
        public static T? TryCast<T>(this object? obj)
        {
            try { return (T?)Convert.ChangeType(obj, typeof(T)); }
            catch { return default; }
        }
    }
}
