using sz84.Bots.BotMessages;
using sz84.Core.MetaDatas;

namespace sz84.Bots.Entries
{
    public class BotEventLog : MetaData<BotEventLog>
    {

        public override string TableName => "Event";
        public override string KeyField => "Id";

        // 记录机器人事件
        public static int Append(BotMessage bm, string eventName)
        {
            return Append(bm.SelfId, eventName, bm.GroupId, bm.GroupName, bm.UserId, bm.Name);
        }

        // 记录机器人事件
        public static int Append(long botUin, string eventName, long groupId, string groupName, long userId, string userName)
        {
            return Insert([
                    new Cov("BotUin", botUin),
                    new Cov("GroupId", groupId),
                    new Cov("GroupName", groupName),
                    new Cov("UserId", userId),
                    new Cov("UserName", userName),
                    new Cov("EventName", eventName),
                ]);
        }
    }
}
