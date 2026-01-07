namespace BotWorker.Infrastructure.Utils.Schema;

public static class TableNameHelper
{
    public static string GetTableName<T>()
    {
        var type = typeof(T);
        return type.Name.ToLower(); // 可替换为带前缀、下划线风格等
    }
}
