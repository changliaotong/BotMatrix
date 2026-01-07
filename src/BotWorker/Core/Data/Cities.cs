using sz84.Core;
using sz84.Core.MetaDatas;

namespace sz84.Core.Data
{
    /// <summary>
    /// 中国城市数据
    /// </summary>
    public class Cities : MetaData<Cities>
    {
        public override string DataBase => "baseinfo";
        public override string TableName => "cities";
        public override string KeyField => "city_name";
    }
}
