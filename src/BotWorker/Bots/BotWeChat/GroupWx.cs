using sz84.Core.MetaDatas;

namespace sz84.Bots.BotWeChat
{
    public class GroupWx : MetaData<GroupWx>
    {
        public override string TableName => "wx_group";
        public override string KeyField => "group_id";
    }
}
