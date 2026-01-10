using System.Data;
using System.Reflection;
using BotWorker.Infrastructure.Persistence.Database;

namespace BotWorker.Infrastructure.Persistence.ORM
{
    public static class EntityMapper
    {
        public static T MapDataReaderToEntity<T>(IDataRecord record) where T : new()
        {
            var entity = new T();
            var props = typeof(T).GetProperties();

            foreach (var prop in props)
            {
                string columnName = SqlHelper.GetColumnName(prop);

                int ordinal = -1;
                try
                {
                    ordinal = record.GetOrdinal(columnName);
                }
                catch (IndexOutOfRangeException)
                {
                    // 数据库返回的列里没这个字段，跳过
                    continue;
                }

                if (record.IsDBNull(ordinal))
                {
                    prop.SetValue(entity, null);
                    continue;
                }

                var dbValue = record.GetValue(ordinal);

                var attr = prop.GetCustomAttribute<ColumnAttribute>();
                if (attr?.ConverterType != null)
                {
                    var converter = Activator.CreateInstance((Type)attr.ConverterType) as IValueConverter;
                    if (converter != null)
                    {
                        var val = converter.ConvertFromProvider(dbValue);
                        prop.SetValue(entity, val);
                        continue;
                    }
                }

                // 不带转换器，直接赋值
                prop.SetValue(entity, dbValue);
            }

            return entity;
        }

    }
}
