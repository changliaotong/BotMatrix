using System.Threading.Tasks;
using System.Data;
using BotWorker.Domain.Entities;

namespace BotWorker.Domain.Repositories
{
    public interface IAnswerRepository : IBaseRepository<AnswerInfo>
    {
        Task<long> AppendAsync(long botUin, long groupId, long qq, long robotId, long questionId, string textQuestion, string textAnswer, int audit, long credit, int audit2, string audit2Info, IDbTransaction? trans = null);
        Task<bool> ExistsAsync(long questionId, string textAnswer, long groupId);
        Task<bool> ExistsAsync(long qqRobot, long questionId, string answer);
        Task<long> CountAnswerAsync(long questionId);
        Task<int> IncrementUsedTimesAsync(long answerId);
        Task<int> AuditAsync(long answerId, int audit, long qq);
    }
}
