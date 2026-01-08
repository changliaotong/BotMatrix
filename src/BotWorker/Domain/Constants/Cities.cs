using BotWorker.Core;
using BotWorker.Core.MetaDatas;

namespace BotWorker.Core.Data
{
    /// <summary>
    /// �й���������
    /// </summary>
    public class Cities : MetaData<Cities>
    {
        public override string DataBase => "baseinfo";
        public override string TableName => "cities";
        public override string KeyField => "city_name";
    }
}


