using sz84.Bots.BotMessages;
using BotWorker.Common.Exts;
using sz84.Core.MetaDatas;
using sz84.Infrastructure.Extensions.Text;

namespace sz84.Agents.Entries
{
    public class AgentLog : MetaData<AgentLog>
    {
        public override string TableName => "AgentLog";
        public override string KeyField => "Id";

        public static int Append(BotMessage bm)
        {
            Agent.UsedTimesIncrement(bm.AgentId);
            return Insert([
                    new Cov("GroupId", bm.RealGroupId),
                    new Cov("Groupname", bm.GroupName),
                    new Cov("UserId", bm.UserId),
                    new Cov("UserName", bm.Name),
                    new Cov("MsgId", bm.MsgId),
                    new Cov("Messages", bm.History.ToJsonString()),
                    new Cov("question", bm.Message),
                    new Cov("answer", bm.AnswerAI),
                    new Cov("AgentId", bm.AgentId),
                    new Cov("ModelId", bm.ModelId),
                    new Cov("TokensInput", bm.InputTokens),
                    new Cov("TokensOutput", bm.OutputTokens),
                    new Cov("TokensTimes", bm.TokensTimes),
                    new Cov("TokensTimesOutput", bm.TokensTimesOutput),
                    new Cov("TokensMinus", bm.TokensMinus),
                    new Cov("CostTime", bm.CurrentStopwatch == null ? 0 : bm.CurrentStopwatch.Elapsed.TotalSeconds)
                ]);
        }
    }

}
