using BotWorker.Core.MetaDatas;

namespace BotWorker.Bots.Public
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

        //�û�����ʼ��
        public static readonly long UinStart = 4104967295;

        //ͨ�� key ��û�����qq
        public static long GetRobotQQ(string botKey)
        {
            return GetLong("BotUin", botKey);
        }

        // ���ں��Ա�Ⱥ��
        public static long GetGroupId(string botKey)
        {
            return GetLong("GroupId", botKey);
        }

        // ���ں�����
        public static string GetBotName(string botKey)
        {
            return GetDef("PublicName", botKey, "[δ֪���ں�]");
        }
    }
}

