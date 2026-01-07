using sz84.Bots.BotMessages;
using sz84.Core.MetaDatas;

namespace sz84.Bots.Groups
{
    public class GroupEvent : MetaData<GroupEvent>
    {
        public override string TableName => "GroupEvent";
        public override string KeyField => "Id";

        public static int Append(BotMessage bm, string eventType, string eventInfo)
        {
            return Append(bm.SelfId, bm.RealGroupId, bm.RealGroupName, bm.UserId, bm.Name, bm.Message, eventType, eventInfo);
        }

        public static int Append(long botUin, long groupId, string groupName, long qq, string name, string message, string eventType, string eventInfo)
        {
            return Insert([
                new Cov("BotUin", botUin),
                new Cov("GroupId", groupId),
                new Cov("GroupName", groupName),
                new Cov("UserId", qq),
                new Cov("UserName", name),
                new Cov("Message", message),
                new Cov("EventType", eventType),
                new Cov("EventInfo", eventInfo),
            ]);
        }
    }
}
