using System.Threading.Tasks;
using BotWorker.Domain.Repositories;

namespace BotWorker.Domain.Entities
{
    public partial class AnswerInfo
    {
        public static int Append(long botUin, long groupId, long qq, long robotId, long questionId, string textQuestion, string textAnswer,
            int audit, long credit, int audit2, string audit2Info)
        {
            return AppendAsync(botUin, groupId, qq, robotId, questionId, textQuestion, textAnswer, audit, credit, audit2, audit2Info).GetAwaiter().GetResult();
        }

        public static async Task<int> AppendAsync(long botUin, long groupId, long qq, long robotId, long questionId, string textQuestion, string textAnswer,
            int audit, long credit, int audit2, string audit2Info)
        {
            await Repository.AppendAsync(botUin, groupId, qq, robotId, questionId, textQuestion, textAnswer, audit, credit, audit2, audit2Info);
            return 1;
        }

        public static async Task<bool> ExistsAsync(long questionId, string textAnswer, long groupId)
        {
            return await Repository.ExistsAsync(questionId, textAnswer, groupId);
        }

        public static async Task<bool> ExistsAsync(long qqRobot, long questionId, string answer)
        {
            return await Repository.ExistsAsync(qqRobot, questionId, answer);
        }

        public static long CountAnswer(long questionId) => CountAnswerAsync(questionId).GetAwaiter().GetResult();
        public static async Task<long> CountAnswerAsync(long questionId) => await Repository.CountAnswerAsync(questionId);

        public static async Task<int> CountUsedPlusAsync(long answerId) => await Repository.IncrementUsedTimesAsync(answerId);

        public static async Task<int> AuditItAsync(long answerId, int audit, long qq) => await Repository.AuditAsync(answerId, audit, qq);
    }
}
