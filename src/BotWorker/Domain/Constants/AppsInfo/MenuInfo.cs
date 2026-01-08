namespace BotWorker.Domain.Constants.AppsInfo
{
    public class MenuInfo : MetaData<MenuInfo>
    {
        public override string DataBase => "apps";
        public override string TableName => "menu_info";
        public override string KeyField => "menu_id";
    }
}


