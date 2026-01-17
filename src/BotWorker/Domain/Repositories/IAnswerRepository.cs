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
        Task<long> GetGroupAnswerIdAsync(long groupId, long questionId, int length = 0);
        Task<long> GetDefaultAnswerIdAsync(long questionId, long robotId, int length = 0);
        Task<long> GetDefaultAnswerAtIdAsync(long questionId, int length = 0);
        Task<long> GetAllAnswerAuditIdAsync(long questionId, int length = 0);
        Task<long> GetAllAnswerNotAuditIdAsync(long questionId, int length = 0);
        Task<long> GetAllAnswerIdAsync(long questionId, int length = 0);
        Task<long> GetStoryIdAsync();
        Task<long> GetGhostStoryIdAsync();
        Task<long> GetCoupletsIdAsync();
        Task<long> GetChouqianIdAsync();
        Task<long> GetJieqianAnswerIdAsync(long groupId, long userId);
        Task<long> GetAnswerIdByParentAsync(long parentId);
        Task<long> GetDatiIdAsync(string keyword);
    }
}
