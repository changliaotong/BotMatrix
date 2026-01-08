using Microsoft.Data.SqlClient;
namespace BotWorker.Infrastructure.Persistence.ORM
{
    public static class ObjectExtensions
    {
        public static Dictionary<string, object?> ToDictionary(this object anonymous)
        {
            if (anonymous == null)
                return [];

            return anonymous.GetType()
                .GetProperties()
                .ToDictionary(
                    prop => prop.Name,
                    prop => prop.GetValue(anonymous, null)
                );
        }

        // 加一个辅助扩展方法：判断列是否存在
        public static bool HasColumn(this SqlDataReader reader, string columnName)
        {
            for (int i = 0; i < reader.FieldCount; i++)
            {
                if (reader.GetName(i).Equals(columnName, StringComparison.OrdinalIgnoreCase))
                    return true;
            }
            return false;
        }
    }

}
