using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Infrastructure.Communication.Platforms.BotWeChat
{
    public class GroupWx : MetaData<GroupWx>
    {
        public override string TableName => "wx_group";
        public override string KeyField => "Id";

        public long Id { get; set; }
        public string NickName { get; set; } = string.Empty;
        public string HeadImgUrl { get; set; } = string.Empty;
        public string UserName { get; set; } = string.Empty;
        public string OwnerUserName { get; set; } = string.Empty;
        public int MemberCount { get; set; }
    }
}
