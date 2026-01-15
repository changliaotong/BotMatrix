using System.Threading.Tasks;
using System.Data;
using BotWorker.Domain.Entities;

namespace BotWorker.Domain.Repositories
{
    public interface ITokensLogRepository : IBaseRepository<TokensLog>
    {
        Task<int> AddLogAsync(long botUin, long groupId, string groupName, long userId, string userName, long tokensAdd, long tokensValue, string tokensInfo, IDbTransaction? trans = null);
    }
}
