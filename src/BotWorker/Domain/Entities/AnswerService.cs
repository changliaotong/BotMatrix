using System.Threading.Tasks;
using BotWorker.Domain.Models.BotMessages;

namespace BotWorker.Domain.Entities
{
    public partial class AnswerInfo
    {
        public static int Append(BotMessage bm, long botUin, long groupId, long qq, long robotId, long questionId, string textQuestion, string textAnswer,
            int audit, long credit, int audit2, string audit2Info)
        {
            return AppendAsync(bm, botUin, groupId, qq, robotId, questionId, textQuestion, textAnswer, audit, credit, audit2, audit2Info).GetAwaiter().GetResult();
        }

        public static async Task<int> AppendAsync(BotMessage bm, long botUin, long groupId, long qq, long robotId, long questionId, string textQuestion, string textAnswer,
            int audit, long credit, int audit2, string audit2Info)
        {
            await bm.AnswerRepository.AppendAsync(botUin, groupId, qq, robotId, questionId, textQuestion, textAnswer, audit, credit, audit2, audit2Info);
            return 1;
        }

        public static async Task<bool> ExistsAsync(BotMessage bm, long questionId, string textAnswer, long groupId)
        {
            return await bm.AnswerRepository.ExistsAsync(questionId, textAnswer, groupId);
        }

        public static async Task<bool> ExistsAsync(BotMessage bm, long qqRobot, long questionId, string answer)
        {
            return await bm.AnswerRepository.ExistsAsync(qqRobot, questionId, answer);
        }

        public static long CountAnswer(BotMessage bm, long questionId) => CountAnswerAsync(bm, questionId).GetAwaiter().GetResult();
        public static async Task<long> CountAnswerAsync(BotMessage bm, long questionId) => await bm.AnswerRepository.CountAnswerAsync(questionId);

        public static async Task<int> CountUsedPlusAsync(BotMessage bm, long answerId) => await bm.AnswerRepository.IncrementUsedTimesAsync(answerId);

        public static async Task<int> AuditItAsync(BotMessage bm, long answerId, int audit, long qq) => await bm.AnswerRepository.AuditAsync(answerId, audit, qq);
    }
}
