using BotWorker.Core.MetaDatas;
using BotWorker.Bots.Entries;
using BotWorker.Bots.Users;

namespace BotWorker.Bots.Groups
{
    public class WhiteList : MetaData<WhiteList>
    {
        //白名单系统
        public override string TableName => "WhiteList";
        public override string KeyField => "GroupId";
        public override string KeyField2 => "WhiteId";


        // 加入白名单
        public static int AppendWhiteList(long botUin, long groupId, string groupName, long qq, string name, long qqWhite)
        {
            return Exists(groupId, qqWhite) 
                ? 0 
                : Insert([
                            new Cov("BotUin", botUin),
                            new Cov("GroupId", groupId),
                            new Cov("GroupName", groupName),
                            new Cov("UserId", qq),
                            new Cov("UserName", name),
                            new Cov("WhiteId", qqWhite),
                        ]);
        }





    }
}
