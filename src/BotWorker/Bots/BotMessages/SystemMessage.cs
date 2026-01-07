using sz84.Bots.Groups;
using BotWorker.Common.Exts;
using sz84.Core.MetaDatas;

namespace sz84.Bots.BotMessages
{
    public partial class BotMessage : MetaData<BotMessage>
    {
        public string GetRecallCount()
        {
            return GroupEvent.CountWhere($"GroupId = {GroupId}").AsString();
        }

        public enum GroupEventType
        {
            撤回,
            禁言,
            踢出,
            拉黑,
            扣分,
            警告,
        }
        public string GetEventCount(GroupEventType eventType)
        {
            return GroupEvent.CountWhere($"GroupId = {GroupId} AND EventType = {eventType.ToString().Quotes()}").AsString();
        }
    }
}
