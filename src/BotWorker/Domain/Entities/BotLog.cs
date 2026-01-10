using BotWorker.Domain.Models.Messages.BotMessages;
using BotWorker.Domain.Constants;
using BotWorker.Common.Extensions;
using BotWorker.Infrastructure.Persistence.ORM;

namespace BotWorker.Domain.Entities
{
    public class BotLog : MetaData<BotLog>
    {
        public override string DataBase => "sz84_log";
        public override string TableName => "bot_logs";
        public override string KeyField => "id";

        public static int Log(string info, string memo, BotMessage? bm = default)
        {
            var boMessage = string.Empty;
            if (bm != null)
            {
                var properties = bm.GetType().GetProperties();
                var propertyStrings = properties.Select(p => $"{p.Name}: {p.GetValue(bm)}").ToArray();
                boMessage = string.Join(", ", propertyStrings);


                return Insert([
                    new Cov("MsgGuid", bm.MsgGuid),
                    new Cov("info", info),
                    new Cov("memo", memo),
                    new Cov("Json", boMessage),
                    new Cov("BotType", Platforms.BotType(bm.Platform)),
                    new Cov("IsSignalR", bm.SelfInfo.IsSignalR),
                    new Cov("EventType", bm.EventType),
                    new Cov("EventMessage", bm.EventMessage),
                    new Cov("BotQQ", bm.SelfId),
                    new Cov("BotName", bm.SelfName),
                    new Cov("GroupId", bm.GroupId),
                    new Cov("GroupName", bm.GroupName),
                    new Cov("GroupOpenid", bm.GroupOpenid),
                    new Cov("QQ", bm.UserId),
                    new Cov("Name", bm.Name),
                    new Cov("UserOpenid", bm.UserOpenId),
                    new Cov("MsgId", bm.MsgId),
                    new Cov("Message", bm.Message),
                    new Cov("Operater", bm.Operater),
                    new Cov("OperaterName", bm.OperaterName),
                    new Cov("InvitorQQ", bm.InvitorQQ),
                    new Cov("InvitorName", bm.InvitorName),
                    new Cov("Period", bm.Period),
                    new Cov("BotPerm", bm.SelfPerm),
                    new Cov("Perm", bm.UserPerm),
                    new Cov("IsAtMe", bm.IsAtMe),
                    new Cov("IsGroup", bm.IsGroup),
                    new Cov("GroupOwner", bm.Group.RobotOwner),
                    new Cov("IsCmd", bm.IsCmd),                    
                    new Cov("IsRefresh", bm.IsRefresh),
                    new Cov("RealGroupId", bm.GroupId),
                    new Cov("RealMessage", bm.Message),
                    new Cov("CmdName", bm.CmdName),
                    new Cov("CmdPara", bm.CmdPara),
                    new Cov("IsConfirm", bm.IsConfirm),
                    new Cov("AgentId", bm.AgentId),
                    new Cov("Context", bm.HistoryMessageCount),
                    new Cov("AgentName", bm.AgentName),
                    new Cov("InputTokens", bm.InputTokens),
                    new Cov("OutputTokens", bm.OutputTokens),
                    new Cov("TokensTimes", bm.TokensTimes),
                    new Cov("TokensTimesOutput", bm.TokensTimesOutput),
                    new Cov("TokensMinus", bm.TokensMinus),
                    new Cov("ModelId", bm.ModelId),
                    new Cov("IsDup", bm.IsDup),
                    new Cov("IsMusic", bm.IsMusic),
                    new Cov("AnswerId", bm.AnswerId),
                    new Cov("Answer", bm.Answer),
                    new Cov("IsAI", bm.IsAI),
                    new Cov("IsSend", bm.IsSend),
                    new Cov("IsRecall", bm.IsRecall),
                    new Cov("TargetUin", bm.TargetUin),
                    new Cov("Accept", bm.Accept),
                    new Cov("Reason", bm.Reason),
                    new Cov("IsSet", bm.IsSet),
            ]);
            }
            else
                return -1;

        }
    }
}
