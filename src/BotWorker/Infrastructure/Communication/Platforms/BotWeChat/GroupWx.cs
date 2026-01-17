namespace BotWorker.Infrastructure.Communication.Platforms.BotWeChat
{
    public class GroupWx
    {
        public long Id { get; set; }
        public string NickName { get; set; } = string.Empty;
        public string HeadImgUrl { get; set; } = string.Empty;
        public string UserName { get; set; } = string.Empty;
        public string OwnerUserName { get; set; } = string.Empty;
        public int MemberCount { get; set; }
    }
}
