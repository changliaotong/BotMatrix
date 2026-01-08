using Microsoft.Data.SqlClient;

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

        public static (string, SqlParameter[]) SqlAppend(long botUin, long groupId, long qq, long robotId, long questionId, string textQuestion, string textAnswer,
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

        public static bool Exists(long questionId, string textAnswer, long groupId) =>
            ExistsWhere($"QuestionId = {questionId} AND RobotId = {groupId} AND {DbName}.dbo.remove_biaodian(answer) = {DbName}.dbo.remove_biaodian({textAnswer.Quotes()})");

        public static bool Exists(long qqRobot, long questionId, string answer) =>
            ExistsWhere($"QuestionId = {questionId} AND {DbName}.dbo.remove_biaodian(answer) = {DbName}.dbo.remove_biaodian({answer.Quotes()})");

        public static long CountAnswer(long questionId) => CountField("Id", "QuestionId", questionId);

        public static int CountUsedPlus(long answerId) => Plus("UsedTimes", 1, answerId);

        public static int AuditIt(long answerId, int audit, long qq) => Update($"audit = {audit}, AuditBy = {qq}, AuditDate = GETDATE()", answerId);

    }
}
