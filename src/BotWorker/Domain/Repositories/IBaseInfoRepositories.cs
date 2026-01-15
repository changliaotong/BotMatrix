using System.Collections.Generic;
using System.Threading.Tasks;
using BotWorker.Domain.Entities;

namespace BotWorker.Domain.Repositories
{
    public interface IChengyuRepository : IBaseRepository<Chengyu>
    {
        Task<long> GetOidAsync(string text);
        Task<Chengyu?> GetByNameAsync(string name);
        Task<string> GetCyInfoAsync(string text, long oid = 0);
        Task<string> GetInfoHtmlAsync(string text, long oid = 0);
        Task<long> CountBySearchAsync(string search);
        Task<string> SearchCysAsync(string search, int top = 50);
        Task<long> GetOidBySearchAsync(string search);
        Task<long> CountByFanChaAsync(string search);
        Task<string> SearchByFanChaAsync(string search, int top = 50);
    }

    public interface ICidianRepository : IBaseRepository<Cidian>
    {
        Task<string> GetDescriptionAsync(string keyword);
        Task<IEnumerable<Cidian>> SearchAsync(string keyword, int limit = 20);
    }

    public interface ICityRepository : IBaseRepository<City>
    {
        Task<City?> GetByNameAsync(string cityName);
    }

    public interface IQuestionInfoRepository : IBaseRepository<QuestionInfo>
    {
        Task<bool> ExistsByQuestionAsync(string question);
        Task<long> AddQuestionAsync(long botUin, long groupId, long userId, string question);
        Task<int> IncrementUsedTimesAsync(long questionId);
        Task<bool> IsSystemAsync(long questionId);
        Task<int> AuditAsync(long questionId, int audit2, int isSystem);
        Task<long> GetIdByQuestionAsync(string question);
    }
}
