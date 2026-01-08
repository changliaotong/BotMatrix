namespace BotWorker.Core.Database.Mapping
{
    public interface IValueConverter
    {
        /// <summary>
        /// 将属性值转换为数据库可接受的值
        /// </summary>
        object? ConvertToProvider(object? value);

        /// <summary>
        /// 将数据库值转换为属性值（反向转换，可选）
        /// </summary>
        object? ConvertFromProvider(object? value);
    }
}
