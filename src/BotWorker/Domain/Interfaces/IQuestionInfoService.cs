using System.Threading.Tasks;

namespace BotWorker.Domain.Interfaces
{
    public interface IQuestionInfoService
    {
        Task<bool> ExistsByQuestionAsync(string question);
        Task<long> AddQuestionAsync(long botUin, long groupId, long userId, string question);
        Task<int> IncrementUsedTimesAsync(long questionId);
        Task<bool> IsSystemAsync(long questionId);
        Task<int> AuditAsync(long questionId, int audit2, int isSystem);
        Task<long> GetIdByQuestionAsync(string question);
        string GetNew(string text);
    }
}
