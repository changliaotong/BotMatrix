namespace BotWorker.Common.Exts
{
    static class EnumExt
    {
        public static string ToDescription(this Enum value)
        {
            return value.GetDescription();
        }

        /// <summary>
        /// 获取枚举值的所有名称和值列表
        /// </summary>
        public static IEnumerable<(int Value, string Name)> GetEnumList<T>() where T : Enum
            => Enum.GetValues(typeof(T)).Cast<T>().Select(e => (Convert.ToInt32(e), e.ToString()));

        /// <summary>
        /// 将整数转换为枚举（不安全，不检查是否定义）
        /// </summary>
        public static T ToEnum<T>(this int value) where T : Enum
            => (T)Enum.ToObject(typeof(T), value);

        /// <summary>
        /// 判断枚举值是否定义于该枚举类型中
        /// </summary>
        public static bool IsDefinedEnumValue<T>(this T value) where T : Enum
            => Enum.IsDefined(typeof(T), value);

        /// <summary>
        /// 获取枚举的描述字符串（需要使用 [Description("xxx")]）
        /// </summary>
        public static string GetDescription<T>(this T enumValue) where T : Enum
        {
            var field = enumValue.GetType().GetField(enumValue.ToString());
            var attr = field?.GetCustomAttributes(typeof(System.ComponentModel.DescriptionAttribute), false)
                             .FirstOrDefault() as System.ComponentModel.DescriptionAttribute;
            return attr?.Description ?? enumValue.ToString();
        }
    }
}
