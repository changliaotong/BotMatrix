namespace BotWorker.Domain.Entities
{
    public class GroupSendMessage : MetaData<GroupSendMessage>
    {
        public override string TableName => "SendMessage";
        public override string KeyField => "Id";
        
        public long MsgId { get; set; }
        public long? GroupId { get; set; }
        public long? ClientQQ { get; set; }
        public string? Question { get; set; }
        public string? AnswerAi { get; set; }
        public DateTime? InsertDate { get; set; }

        public static int UserCount(long groupId)
        {
            return QueryScalar<int>($"SELECT COUNT(DISTINCT UserId) FROM {FullName} WHERE DATEDIFF(SECOND, InsertDate, GETDATE()) < 60 AND GroupId = {groupId}");
        }

        public static int Append(BotMessage bm)
        {
            if (bm.User.IsLog) BotLog.Log($"{bm.GroupName}({bm.GroupId}) {bm.Name}({bm.UserId}) {bm.EventType}：\n{bm.Message}", "处理后", bm);
            return !(bm.IsBlackSystem && bm.EventType.In("EventPrivateMessage", "EventGroupMessage", "TempMessageEvent")) ?
                Insert([                           
                            new Cov("MsgGuid", bm.MsgGuid),
                            new Cov("BotUin", bm.SelfId),
                            new Cov("GroupId", bm.RealGroupId),
                            new Cov("GroupName", bm.RealGroupName),
                            new Cov("UserId", bm.UserId),
                            new Cov("UserName", bm.Name),
                            new Cov("MsgId", bm.MsgId),
                            new Cov("Question", bm.Message.IsNull() ? bm.EventType : bm.Message),
                            new Cov("Message", bm.IsSend && bm.IsRealProxy && !bm.IsAI && bm.AnswerId == 0 ? $"@{bm.Card.ReplaceInvalid().RemoveUserIds().ReplaceSensitive(Regexs.OfficalRejectWords)}:{bm.Answer}" : bm.Answer),
                            new Cov("AnswerAI", bm.AnswerAI),
                            new Cov("AnswerId", bm.AnswerId),
                            new Cov("IsAI", bm.IsAI),
                            new Cov("AgentId", bm.AgentId),
                            new Cov("AgentName", bm.AgentName),
                            new Cov("IsSend", bm.IsSend),
                            new Cov("IsRealProxy", bm.IsRealProxy),
                            new Cov("Reason", bm.Reason),
                            new Cov("IsCmd", bm.IsCmd),
                            new Cov("InputTokens", bm.InputTokens),
                            new Cov("OutputTokens", bm.OutputTokens),
                            new Cov("TokensMinus", bm.TokensMinus),
                            new Cov("IsVoiceReply", bm.IsVoiceReply),
                            new Cov("VoiceName", bm.VoiceName),
                            new Cov("CostTime", bm.CostTime),
                            new Cov("IsRecall", bm.IsRecall),
                            new Cov("ReCallAfterMs", bm.RecallAfterMs),
                        ])
                : 0;
        }
    }
}
