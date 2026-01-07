namespace BotWorker.Common.Exts
{

    public static class FluentExtensions
    {
        // 支持链式配置
        public static T Tap<T>(this T obj, Action<T> action)
        {
            action?.Invoke(obj);
            return obj;
        }

        // 条件执行
        public static T When<T>(this T obj, bool condition, Action<T> action)
        {
            if (condition) action(obj);
            return obj;
        }
    }
}
