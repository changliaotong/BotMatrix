using System.Data;

namespace BotWorker.Domain.Entities
{
    public partial class AnswerInfo : MetaDataGuid<AnswerInfo>
    {
        public static int Append(long botUin, long groupId, long qq, long robotId, long questionId, string textQuestion, string textAnswer,
            int audit, long credit, int audit2, string audit2Info)
        {
            var (sql, parameters) = SqlAppend(botUin, groupId, qq, robotId, questionId, textQuestion, textAnswer, audit, credit, audit2, audit2Info);
            return Exec(sql, parameters);
        }

        public static async Task<int> AppendAsync(long botUin, long groupId, long qq, long robotId, long questionId, string textQuestion, string textAnswer,
            int audit, long credit, int audit2, string audit2Info)
        {
            var (sql, parameters) = SqlAppend(botUin, groupId, qq, robotId, questionId, textQuestion, textAnswer, audit, credit, audit2, audit2Info);
            return await ExecAsync(sql, parameters);
        }

        public static (string, IDataParameter[]) SqlAppend(long botUin, long groupId, long qq, long robotId, long questionId, string textQuestion, string textAnswer,
            int audit, long credit, int audit2, string audit2Info)
        {
            return SqlInsert([
                                new Cov("BotUin", botUin),
                                new Cov("GroupId", groupId),
                                new Cov("UserId", qq),
                                new Cov("RobotId", robotId),
                                new Cov("QuestionId", questionId),
                                new Cov("Question", textQuestion),
                                new Cov("Answer", textAnswer),
                                new Cov("Audit", audit),
                                new Cov("Credit", credit),
                                new Cov("Audit2", audit2),
                                new Cov("Audit2Info", audit2Info)
                        ]);
        }

        public static async Task<bool> ExistsAsync(long questionId, string textAnswer, long groupId)
        {
            string func = IsPostgreSql ? "remove_biaodian" : $"{DbName}.dbo.remove_biaodian";
            return await ExistsWhereAsync($"QuestionId = {questionId} AND RobotId = {groupId} AND {func}(answer) = {func}({textAnswer.Quotes()})");
        }

        public static async Task<bool> ExistsAsync(long qqRobot, long questionId, string answer)
        {
            string func = IsPostgreSql ? "remove_biaodian" : $"{DbName}.dbo.remove_biaodian";
            return await ExistsWhereAsync($"QuestionId = {questionId} AND {func}(answer) = {func}({answer.Quotes()})");
        }

        public static long CountAnswer(long questionId) => CountAnswerAsync(questionId).GetAwaiter().GetResult();
        public static async Task<long> CountAnswerAsync(long questionId) => await CountFieldAsync("Id", "QuestionId", questionId);

        public static async Task<int> CountUsedPlusAsync(long answerId) => await PlusAsync("UsedTimes", 1, answerId);

        public static async Task<int> AuditItAsync(long answerId, int audit, long qq) => await UpdateAsync($"audit = {audit}, AuditBy = {qq}, AuditDate = {SqlDateTime}", answerId);

    }
}
