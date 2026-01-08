using BotWorker.Core;
using BotWorker.Core.MetaDatas;

namespace BotWorker.Core.Data.AppsInfo
{
    public class MenuInfo : MetaData<MenuInfo>
    {
        public override string DataBase => "apps";
        public override string TableName => "menu_info";
        public override string KeyField => "menu_id";
    }
}


