namespace BotWorker.Infrastructure.Communication.Platforms.BotWeChat
{
    public class ClientWx
    {
        public long Id { get; set; }
        public string NickName { get; set; } = string.Empty;
        public string HeadImgUrl { get; set; } = string.Empty;
        public string UserName { get; set; } = string.Empty;
        public string Alias { get; set; } = string.Empty;
        public int Sex { get; set; }
        public string Signature { get; set; } = string.Empty;
        public string Province { get; set; } = string.Empty;
        public string City { get; set; } = string.Empty;
    }
}
