using BotWorker.common.Exts;

namespace sz84.Core.Database.Mapping
{
    public class EnumValueConverter<TEnum> : IValueConverter where TEnum : struct, Enum
    {
        // 从数据库值 → 枚举属性值
        public object? Convert(object? dbValue)
        {
            if (dbValue == null || dbValue == DBNull.Value)
                return null;

            // 支持数字或字符串枚举表示
            if (dbValue is string s)
            {
                return Enum.TryParse<TEnum>(s, ignoreCase: true, out var result) ? result : null;
            }

            if (dbValue is int || dbValue is byte || dbValue is short || dbValue is long)
            {
                return (TEnum)Enum.ToObject(typeof(TEnum), dbValue);
            }

            return null;
        }

        // 从枚举属性值 → 数据库值
        public object? ConvertToProvider(object? value)
        {
            if (value == null)
                return DBNull.Value;

            return Convert(value).AsInt(); 
        }

        // 这里保留接口一致性，但你用不到时可以忽略这个
        public object? ConvertFromProvider(object? value)
        {
            return Convert(value);
        }
    }

}
