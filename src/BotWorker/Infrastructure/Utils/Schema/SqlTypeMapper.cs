namespace BotWorker.Infrastructure.Utils.Schema;

public static class SqlTypeMapper
{
    public static string Map(Type type, bool isPrimaryKey = false)
    {
        var t = Nullable.GetUnderlyingType(type) ?? type;

        return t switch
        {
            var x when x == typeof(int) => "INT",
            var x when x == typeof(long) => "BIGINT",
            var x when x == typeof(Guid) => "UNIQUEIDENTIFIER",
            var x when x == typeof(string) => isPrimaryKey ? "NVARCHAR(255)" : "NVARCHAR(MAX)",
            var x when x == typeof(bool) => "BIT",
            var x when x == typeof(DateTime) => "DATETIME",
            var x when x == typeof(float) || x == typeof(double) => "FLOAT",
            var x when x == typeof(decimal) => "DECIMAL(18,2)",
            _ => isPrimaryKey ? "NVARCHAR(255)" : "NVARCHAR(MAX)" // 默认 fallback
        };
    }
}
