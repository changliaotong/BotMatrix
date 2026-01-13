namespace BotWorker.Domain.Entities
{
    public class MusicAnswer : AnswerInfo
    {
        public static void Append(long botUin, long groupId, long qq, string title, string songId, string answer)
        {
            var questionId = QuestionInfo.GetQId(title);
            if (questionId == 0)
                questionId = QuestionInfo.Append(botUin, groupId, qq, title);
            if (CountAnswer(questionId) == 0)
            {
                if (!ExistsAandB("Id", questionId, "Answer", answer))
                {
                    if (Append(botUin, groupId, qq, BotInfo.GroupCrm, questionId, title, answer, 1, 0, 2, "系统问答") == -1)
                        Logger.Error($"添加答案失败：{title} {songId}");
                    else
                    {
                        QuestionInfo.Audit(questionId, 1, 1);
                        Logger.Info($"✅ 已添加：Title:{title} SongId:{songId}");
                    }
                }
            }
        }

        public static (long, string) RandomMusic()
        {
            var answerId = QueryScalar<long>($"SELECT TOP 1 Id FROM {FullName} WHERE question LIKE '%点歌%' " +
                                 $"AND audit  = 1 AND audit2 > 2 AND UsedTimes > 5 AND GoonTimes > 3 AND len(answer) > 20 " +
                                 $"ORDER BY NEWID()");
            var answer = GetValue("answer", answerId);
            return (answerId, answer);
        }
    }
}
