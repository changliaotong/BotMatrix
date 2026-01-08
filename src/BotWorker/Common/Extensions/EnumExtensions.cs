namespace BotWorker.BotWorker.BotWorker.Common.Exts
{
    static class EnumExt
    {
        public static string ToDescription(this Enum value)
        {
            return value.GetDescription();
        }

        /// <summary>
        /// ��ȡö��ֵ���������ƺ�ֵ�б�
        /// </summary>
        public static IEnumerable<(int Value, string Name)> GetEnumList<T>() where T : Enum
            => Enum.GetValues(typeof(T)).Cast<T>().Select(e => (Convert.ToInt32(e), e.ToString()));

        /// <summary>
        /// ������ת��Ϊö�٣�����ȫ��������Ƿ��壩
        /// </summary>
        public static T ToEnum<T>(this int value) where T : Enum
            => (T)Enum.ToObject(typeof(T), value);

        /// <summary>
        /// �ж�ö��ֵ�Ƿ����ڸ�ö��������
        /// </summary>
        public static bool IsDefinedEnumValue<T>(this T value) where T : Enum
            => Enum.IsDefined(typeof(T), value);

        /// <summary>
        /// ��ȡö�ٵ������ַ�������Ҫʹ�� [Description("xxx")]��
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


