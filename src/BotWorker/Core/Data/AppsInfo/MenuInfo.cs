using sz84.Core;
using sz84.Core.MetaDatas;

namespace sz84.Core.Data.AppsInfo
{
    public class MenuInfo : MetaData<MenuInfo>
    {
        public override string DataBase => "apps";
        public override string TableName => "menu_info";
        public override string KeyField => "menu_id";
    }
}
