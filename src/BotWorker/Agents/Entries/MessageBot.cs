using Newtonsoft.Json;
using BotWorker.Bots.Entries;
using BotWorker.Common.Exts;

namespace BotWorker.Agents.Entries
{
    public class MessageBot
    {
        [JsonProperty("Id")] 
        public long MsgId { get; set; }

        [JsonProperty(nameof(BotUin))] 
        public long BotUin { get; set; }
        public string RobotName => BotInfo.GetValue("BotName", BotUin);

        [JsonProperty(nameof(GroupId))]
        public long GroupId { get; set; }

        [JsonProperty(nameof(GroupName))]
        public string GroupName { get; set; } = string.Empty;

        [JsonProperty("UserId")]
        public long UserId { get; set; }

        [JsonProperty(nameof(UserName))]
        public string UserName { get; set; } = string.Empty;

        [JsonProperty("InsertDate")]
        public string SendTime { get; set; } = string.Empty;

        [JsonProperty(nameof(IsAI))]
        public bool IsAI { get; set; }

        [JsonProperty(nameof(Question))]
        public string Question { get; set; } = string.Empty;

        [JsonProperty(nameof(Message))]
        public string Message { get; set; } = string.Empty;

        public static UserMessage GetUserMessage(MessageBot botMsg, bool isRobot, long qq)
        {
            return new UserMessage
            {
                MsgId = botMsg.MsgId,
                RobotQQ = botMsg.BotUin,
                GroupId = botMsg.GroupId,
                GroupName = botMsg.GroupName,
                QQ = isRobot ? botMsg.BotUin : botMsg.UserId,
                ClientName = isRobot ? botMsg.GroupName.IsNull()? botMsg.RobotName : botMsg.GroupName : botMsg.UserName,
                SendTime = botMsg.SendTime,
                IsCurr = !isRobot && qq == botMsg.UserId,
                IsAI = botMsg.IsAI,
                Message = isRobot ? botMsg.Message : botMsg.Question,
            };
        }
    }
}
