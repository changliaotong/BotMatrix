using sz84.Core.MetaDatas;

namespace sz84.Bots.Public
{
    public class BotPublic : MetaData<BotPublic>
    {
        public override string TableName => "Public";
        public override string KeyField => "PublicKey";

        public string PublicKey { get; set; } = string.Empty;        
        public string PublicName {  get; set; } = string.Empty;
        public long GroupId { get; set; }
        public long BotUin { get; set; }
        public long AdminId { get; set; }

        //用户号起始段
        public static readonly long UinStart = 4104967295;

        //通过 key 获得机器人qq
        public static long GetRobotQQ(string botKey)
        {
            return GetLong("BotUin", botKey);
        }

        // 公众号自编群号
        public static long GetGroupId(string botKey)
        {
            return GetLong("GroupId", botKey);
        }

        // 公众号名称
        public static string GetBotName(string botKey)
        {
            return GetDef("PublicName", botKey, "[未知公众号]");
        }
    }
}