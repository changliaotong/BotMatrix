namespace BotWorker.BotWorker.BotWorker.Common.Exts
{
    public static class TypeExtensions
    {
        // �ж϶����Ƿ��ת��Ϊ����
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


