namespace BotWorker.BotWorker.BotWorker.Common.Exts
{

    public static class FluentExtensions
    {
        // ֧����ʽ����
        public static T Tap<T>(this T obj, Action<T> action)
        {
            action?.Invoke(obj);
            return obj;
        }

        // ����ִ��
        public static T When<T>(this T obj, bool condition, Action<T> action)
        {
            if (condition) action(obj);
            return obj;
        }
    }
}


