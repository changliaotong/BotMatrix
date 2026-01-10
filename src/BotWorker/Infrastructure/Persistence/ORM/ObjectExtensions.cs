using System.Data;
using BotWorker.Infrastructure.Persistence.Database;

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

        public static IDataParameter[] ToParameters(this object? anonymous)
        {
            if (anonymous == null) return Array.Empty<IDataParameter>();
            if (anonymous is IDataParameter[] paras) return paras;
            if (anonymous is IEnumerable<IDataParameter> enumerable) return enumerable.ToArray();

            var dict = ToDictionary(anonymous);
            return dict.Select(kv => MetaData.CreateParameter($"@{kv.Key}", kv.Value ?? DBNull.Value)).ToArray();
        }
    }
}
